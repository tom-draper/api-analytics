namespace Analytics;

using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using System.Collections.Concurrent;
using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

public class Config
{
    public Func<HttpContext, string>? GetPath { get; set; } = null;
    public Func<HttpContext, string>? GetIPAddress { get; set; } = null;
    public Func<HttpContext, string>? GetHostname { get; set; } = null;
    public Func<HttpContext, string>? GetUserAgent { get; set; } = null;
    public Func<HttpContext, string>? GetUserID { get; set; } = null;
    public int PrivacyLevel { get; set; } = 0;
    public string ServerUrl { get; set; } = "https://www.apianalytics-server.com/api/log-request";
}

public class AnalyticsService : IDisposable
{
    private readonly string _apiKey;
    private readonly Config _config;
    private readonly ILogger<AnalyticsService> _logger;
    private readonly HttpClient _httpClient;
    private readonly ConcurrentQueue<RequestData> _requests = new();
    private readonly Timer _flushTimer;
    private readonly SemaphoreSlim _flushSemaphore = new(1, 1);
    private bool _disposed = false;
    private static readonly TimeSpan FlushInterval = TimeSpan.FromSeconds(60);
    private static readonly TimeSpan HttpTimeout = TimeSpan.FromMinutes(5);

    public AnalyticsService(string apiKey, Config config, ILogger<AnalyticsService> logger)
    {
        _apiKey = apiKey ?? throw new ArgumentNullException(nameof(apiKey));
        _config = config ?? throw new ArgumentNullException(nameof(config));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));

        _httpClient = new HttpClient()
        {
            Timeout = HttpTimeout
        };

        _flushTimer = new Timer(async _ => await FlushRequestsAsync(), null, FlushInterval, FlushInterval);
    }

    private struct Payload
    {
        [JsonPropertyName("api_key")]
        public string ApiKey { get; set; }

        [JsonPropertyName("requests")]
        public List<RequestData> Requests { get; set; }

        [JsonPropertyName("framework")]
        public string Framework { get; set; }

        [JsonPropertyName("privacy_level")]
        public int PrivacyLevel { get; set; }
    }

    public struct RequestData
    {
        [JsonPropertyName("hostname")]
        public string Hostname { get; set; }

        [JsonPropertyName("ip_address")]
        public string IPAddress { get; set; }

        [JsonPropertyName("user_agent")]
        public string UserAgent { get; set; }

        [JsonPropertyName("path")]
        public string Path { get; set; }

        [JsonPropertyName("method")]
        public string Method { get; set; }

        [JsonPropertyName("response_time")]
        public int ResponseTime { get; set; }

        [JsonPropertyName("status")]
        public int Status { get; set; }

        [JsonPropertyName("user_id")]
        public string UserID { get; set; }

        [JsonPropertyName("created_at")]
        public string CreatedAt { get; set; }

        public override readonly string ToString()
        {
            return JsonSerializer.Serialize(this);
        }
    }

    private async Task PostRequestsAsync(List<RequestData> requests)
    {
        if (string.IsNullOrEmpty(_apiKey) || requests.Count == 0)
            return;

        var payload = new Payload
        {
            ApiKey = _apiKey,
            Requests = requests,
            Framework = "ASP.NET Core",
            PrivacyLevel = _config.PrivacyLevel
        };

        try
        {
            var json = JsonSerializer.Serialize(payload);
            using var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await _httpClient.PostAsync(_config.ServerUrl, content);

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogWarning("Failed to post analytics data. Status: {StatusCode}, Reason: {ReasonPhrase}",
                    response.StatusCode, response.ReasonPhrase);
            }
            else
            {
                _logger.LogDebug("Successfully posted {Count} analytics requests", requests.Count);
            }
        }
        catch (TaskCanceledException ex) when (ex.InnerException is TimeoutException)
        {
            _logger.LogWarning("Timeout occurred while posting analytics data: {Message}", ex.Message);
        }
        catch (HttpRequestException ex)
        {
            _logger.LogWarning("HTTP error occurred while posting analytics data: {Message}", ex.Message);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Unexpected error occurred while posting analytics data");
        }
    }

    private async Task FlushRequestsAsync()
    {
        if (!_flushSemaphore.Wait(0)) // Non-blocking wait
            return; // A flush is already in progress, skip this tick.

        try
        {
            var requestsToFlush = new List<RequestData>();
            while (_requests.TryDequeue(out var request))
                requestsToFlush.Add(request);

            if (requestsToFlush.Count > 0)
                await PostRequestsAsync(requestsToFlush);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error occurred during timed flush of analytics data.");
        }
        finally
        {
            _flushSemaphore.Release();
        }
    }

    public void LogRequest(RequestData requestData)
    {
        if (string.IsNullOrEmpty(_apiKey))
            return;

        _requests.Enqueue(requestData);
        // No immediate flush - only timer-based flushing every 60 seconds
    }

    public RequestData CreateRequestData(HttpContext context, long responseTimeMs, DateTime createdAt)
    {
        return new RequestData
        {
            Hostname = GetHostname(context),
            IPAddress = GetIPAddress(context),
            UserAgent = GetUserAgent(context),
            Path = GetPath(context),
            Method = context.Request.Method,
            ResponseTime = (int)responseTimeMs,
            Status = context.Response.StatusCode,
            UserID = GetUserID(context),
            CreatedAt = createdAt.ToString("yyyy-MM-dd'T'HH:mm:ss.fffK")
        };
    }

    private string GetIPAddress(HttpContext context)
    {
        if (_config.PrivacyLevel >= 2)
            return "";

        try
        {
            if (_config.GetIPAddress != null)
                return _config.GetIPAddress.Invoke(context) ?? "";

            // Check for forwarded IP addresses first
            var forwardedFor = context.Request.Headers["X-Forwarded-For"].FirstOrDefault();
            if (!string.IsNullOrEmpty(forwardedFor))
            {
                // Take the first IP if multiple are present
                var firstIp = forwardedFor.Split(',')[0].Trim();
                if (!string.IsNullOrEmpty(firstIp))
                    return firstIp;
            }

            return context.Connection.RemoteIpAddress?.ToString() ?? "";
        }
        catch
        {
            return "";
        }
    }

    private string GetHostname(HttpContext context)
    {
        try
        {
            if (_config.GetHostname != null)
                return _config.GetHostname.Invoke(context) ?? "";
            return context.Request.Host.ToString();
        }
        catch
        {
            return "";
        }
    }

    private string GetUserAgent(HttpContext context)
    {
        try
        {
            if (_config.GetUserAgent != null)
                return _config.GetUserAgent.Invoke(context) ?? "";
            return context.Request.Headers.UserAgent.ToString();
        }
        catch
        {
            return "";
        }
    }

    private string GetPath(HttpContext context)
    {
        try
        {
            if (_config.GetPath != null)
                return _config.GetPath.Invoke(context) ?? "";
            return context.Request.Path.ToString();
        }
        catch
        {
            return "";
        }
    }

    private string GetUserID(HttpContext context)
    {
        try
        {
            if (_config.GetUserID != null)
                return _config.GetUserID.Invoke(context) ?? "";
            return "";
        }
        catch
        {
            return "";
        }
    }

    public async Task FlushAsync()
    {
        await _flushSemaphore.WaitAsync();
        try
        {
            var allRequests = new List<RequestData>();
            while (_requests.TryDequeue(out var request))
                allRequests.Add(request);

            if (allRequests.Count > 0)
                await PostRequestsAsync(allRequests);
        }
        finally
        {
            _flushSemaphore.Release();
        }
    }

    public void Dispose()
    {
        if (!_disposed)
        {
            _flushTimer?.Dispose();

            // Flush any remaining requests synchronously
            try
            {
                FlushAsync().GetAwaiter().GetResult();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error occurred while flushing analytics data during disposal");
            }

            _httpClient?.Dispose();
            _flushSemaphore?.Dispose();
            _disposed = true;
        }
    }
}

public class AnalyticsMiddleware(RequestDelegate next, AnalyticsService analyticsService, ILogger<AnalyticsMiddleware> logger)
{
    private readonly RequestDelegate _next = next ?? throw new ArgumentNullException(nameof(next));
    private readonly AnalyticsService _analyticsService = analyticsService ?? throw new ArgumentNullException(nameof(analyticsService));
    private readonly ILogger<AnalyticsMiddleware> _logger = logger ?? throw new ArgumentNullException(nameof(logger));

    public async Task InvokeAsync(HttpContext context)
    {
        var watch = Stopwatch.StartNew();
        var createdAt = DateTime.UtcNow;

        try
        {
            await _next(context);
        }
        finally
        {
            watch.Stop();

            try
            {
                var requestData = _analyticsService.CreateRequestData(context, watch.ElapsedMilliseconds, createdAt);
                _analyticsService.LogRequest(requestData);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error occurred while logging analytics data");
            }
        }
    }
}

public static class AnalyticsExtensions
{
    public static IApplicationBuilder UseAnalytics(this IApplicationBuilder app, string apiKey, Config? config = null)
    {
        if (string.IsNullOrEmpty(apiKey))
            throw new ArgumentException("API key cannot be null or empty", nameof(apiKey));

        // Create the analytics service
        var loggerFactory = app.ApplicationServices.GetRequiredService<ILoggerFactory>();
        var logger = loggerFactory.CreateLogger<AnalyticsService>();
        var analyticsService = new AnalyticsService(apiKey, config ?? new Config(), logger);

        // Handle disposal during application shutdown
        var appLifetime = app.ApplicationServices.GetRequiredService<IHostApplicationLifetime>();
        appLifetime.ApplicationStopping.Register(() =>
        {
            try
            {
                logger.LogInformation("Analytics service is stopping, flushing remaining data...");
                analyticsService.FlushAsync().GetAwaiter().GetResult();
                analyticsService.Dispose();
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Error occurred while disposing analytics service during shutdown");
            }
        });

        return app.Use(async (context, next) =>
        {
            var middlewareLogger = loggerFactory.CreateLogger<AnalyticsMiddleware>();
            var middleware = new AnalyticsMiddleware(
                async (ctx) => await next(),
                analyticsService,
                middlewareLogger
            );
            await middleware.InvokeAsync(context);
        });
    }
}

public class AnalyticsBackgroundService(AnalyticsService analyticsService, ILogger<AnalyticsBackgroundService> logger) : BackgroundService
{
    private readonly AnalyticsService _analyticsService = analyticsService ?? throw new ArgumentNullException(nameof(analyticsService));
    private readonly ILogger<AnalyticsBackgroundService> _logger = logger ?? throw new ArgumentNullException(nameof(logger));

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        // This service mainly exists to ensure proper cleanup during shutdown
        try
        {
            await Task.Delay(Timeout.Infinite, stoppingToken);
        }
        catch (OperationCanceledException)
        {
            // Expected when cancellation is requested
        }
    }

    public override async Task StopAsync(CancellationToken cancellationToken)
    {
        _logger.LogInformation("Analytics service is stopping, flushing remaining data...");
        try
        {
            await _analyticsService.FlushAsync();
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error occurred while flushing analytics data during shutdown");
        }
        await base.StopAsync(cancellationToken);
    }
}