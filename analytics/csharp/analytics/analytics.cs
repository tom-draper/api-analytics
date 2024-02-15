using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;

public class APIAnalytics
{
    private readonly RequestDelegate _next;
    private static string? _apiKey;
    private DateTime _lastPosted;
    private List<RequestData> _requests;
    private static readonly string _url = "https://www.apianalytics-server.com/api/log-request";

    private struct Payload
    {
        [JsonPropertyName("api_key")]
        public string ApiKey { get; set; }
        [JsonPropertyName("requests")]
        public List<RequestData> Requests { get; set; }
        [JsonPropertyName("framework")]
        public string Framework { get; set; }
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
        [JsonPropertyName("created_at")]

        public string CreatedAt { get; set; }
    }

    public APIAnalytics(RequestDelegate next, string apiKey)
    {
        _next = next;
        _apiKey = apiKey;
        _lastPosted = DateTime.Now;
        _requests = new List<RequestData>();
    }
    
    private static async Task PostRequest(string apiKey, List<RequestData> requests, string framework)
    {
        var payload = new Payload
        {
            ApiKey = apiKey,
            Requests = requests,
            Framework = framework
        };

        var json = JsonSerializer.Serialize(payload);
        var data = new StringContent(json, Encoding.UTF8, "application/json");
        var client = new HttpClient();
        await client.PostAsync(_url, data);
    }

    private async Task LogRequest(RequestData RequestData)
    {
        if (_apiKey == null)
        {
            return;
        }
        var now = DateTime.Now;
        _requests.Add(RequestData);
        if (now.Subtract(_lastPosted).TotalSeconds > 60)
        {
            Thread thread = new Thread(new ThreadStart(async () => await PostRequest(_apiKey, _requests, "ASP.NET Core")));
            await PostRequest(_apiKey, _requests, "ASP.NET Core");
            _requests = new List<RequestData>();
            _lastPosted = now;
        }
    }

    public async Task InvokeAsync(HttpContext context)
    {
        var watch = System.Diagnostics.Stopwatch.StartNew();
        var createdAt = DateTime.Now;
        await _next(context);
        watch.Stop();
        
        var requestData = new RequestData
        {
            Hostname = context.Request.Host.ToString(),
            IPAddress = context.Connection.RemoteIpAddress?.ToString() ?? "",
            UserAgent = context.Request.Headers["User-Agent"].ToString(),
            Path = context.Request.Path.ToString(),
            Method = context.Request.Method.ToString(),
            ResponseTime = watch.Elapsed.Milliseconds,
            Status = context.Response.StatusCode,
            CreatedAt = createdAt.ToUniversalTime().ToString("yyyy-MM-dd'T'HH:mm:ss.fffK")
        };
        await LogRequest(requestData);
    }
}

public static class APIAnalyticsExtensions
{
    public static IApplicationBuilder UseAPIAnalytics(this IApplicationBuilder app, string apiKey)
    {
        return app.UseMiddleware<APIAnalytics>(apiKey);
    }
}
