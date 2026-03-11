using Analytics;

DotNetEnv.Env.Load();
var apiKey = DotNetEnv.Env.GetString("API_KEY");

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.UseAnalytics(apiKey);

app.MapGet("/", () => Results.Ok(new { message = "Hello, World!" }));

app.Run();
