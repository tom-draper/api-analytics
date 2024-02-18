namespace Analytics;

using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;

public class Config
{
    public Func<HttpContext, string>? GetPath { get; set; } = null;
    public Func<HttpContext, string>? GetIPAddress { get; set; } = null;
    public Func<HttpContext, string>? GetHostname { get; set; } = null;
    public Func<HttpContext, string>? GetUserAgent { get; set; } = null;
    public Func<HttpContext, string>? GetUserID { get; set; } = null;
    public int PrivacyLevel { get; set; } = 0;
}

public class Analytics(RequestDelegate next, string apiKey, Config? config = null)
{
    private readonly RequestDelegate _next = next;
    private readonly string? _apiKey = apiKey;
    private readonly Config _config = config ?? new Config();
    private DateTime _lastPosted = DateTime.Now;
    private List<RequestData> _requests = [];
    private readonly string _url = "https://www.apianalytics-server.com/api/log-request";

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

    private async Task PostRequest(string apiKey, List<RequestData> requests, string framework, int privacyLevel)
    {
        var payload = new Payload
        {
            ApiKey = apiKey,
            Requests = requests,
            Framework = framework,
            PrivacyLevel = privacyLevel
        };

        var json = JsonSerializer.Serialize(payload);
        var data = new StringContent(json, Encoding.UTF8, "application/json");
        var client = new HttpClient();
        await client.PostAsync(_url, data);
    }

    private async Task LogRequest(RequestData RequestData)
    {
        if (_apiKey == null)
            return;

        var now = DateTime.Now;
        _requests.Add(RequestData);
        if (now.Subtract(_lastPosted).TotalSeconds > 60)
        {
            Thread thread = new(new ThreadStart(async () => await PostRequest(_apiKey, _requests, "ASP.NET Core", _config.PrivacyLevel)));
            await PostRequest(_apiKey, _requests, "ASP.NET Core", _config.PrivacyLevel);
            _requests = [];
            _lastPosted = now;
        }
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var watch = Stopwatch.StartNew();
        var createdAt = DateTime.Now;
        await _next(context);
        watch.Stop();

        var requestData = new RequestData
        {
            Hostname = GetHostname(context),
            IPAddress = GetIPAddress(context),
            UserAgent = GetUserAgent(context),
            Path = GetPath(context),
            Method = context.Request.Method,
            ResponseTime = watch.Elapsed.Milliseconds,
            Status = context.Response.StatusCode,
            UserID = GetUserID(context),
            CreatedAt = createdAt.ToUniversalTime().ToString("yyyy-MM-dd'T'HH:mm:ss.fffK")
        };
        await LogRequest(requestData);
    }

    private string GetIPAddress(HttpContext context)
    {
        if (_config.PrivacyLevel >= 2)
            return "";

        if (_config.GetIPAddress != null)
            return _config.GetIPAddress.Invoke(context);
        return context.Connection.RemoteIpAddress?.ToString() ?? "";
    }

    private string GetHostname(HttpContext context)
    {
        if (_config.GetHostname != null)
            return _config.GetHostname.Invoke(context);
        return context.Request.Host.ToString();
    }

    private string GetUserAgent(HttpContext context)
    {
        if (_config.GetUserAgent != null)
            return _config.GetUserAgent.Invoke(context) ?? "";
        return context.Request.Headers.UserAgent.ToString();
    }

    private string GetPath(HttpContext context)
    {
        if (_config.GetPath != null)
            return _config.GetPath.Invoke(context);
        return context.Request.Path.ToString();
    }

    private string GetUserID(HttpContext context)
    {
        if (_config.GetUserID != null)
            return _config.GetUserID.Invoke(context);
        return "";
    }
}

public static class AnalyticsExtensions
{
    public static IApplicationBuilder UseAnalytics(this IApplicationBuilder app, string apiKey)
    {
        return app.UseMiddleware<Analytics>(apiKey);
    }
}
