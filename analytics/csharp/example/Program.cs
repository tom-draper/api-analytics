using Microsoft.AspNetCore.Mvc;
using analytics;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.UseAnalytics("");

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
