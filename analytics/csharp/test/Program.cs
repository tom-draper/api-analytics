
using Analytics;
using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

DotNetEnv.Env.Load();
var apiKey = DotNetEnv.Env.GetString("API_KEY");

app.UseAnalytics(apiKey);

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
