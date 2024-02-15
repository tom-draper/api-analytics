using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
