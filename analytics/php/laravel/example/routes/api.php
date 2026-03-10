<?php

/**
 * Laravel Example for API Analytics
 *
 * This file demonstrates how to use API Analytics in a Laravel application.
 *
 * Setup:
 * 1. Add API_ANALYTICS_KEY=your-api-key-here to your .env file
 * 2. Register the middleware in app/Http/Kernel.php (see below)
 * 3. Define your routes as normal
 *
 * Middleware registration in app/Http/Kernel.php:
 *
 *     protected $middleware = [
 *         // ... other middleware
 *         \ApiAnalytics\Laravel\AnalyticsMiddleware::class,
 *     ];
 *
 * Or for API routes only:
 *
 *     protected $middlewareGroups = [
 *         'api' => [
 *             // ... other middleware
 *             \ApiAnalytics\Laravel\AnalyticsMiddleware::class,
 *         ],
 *     ];
 */

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return response()->json(['message' => 'Hello, World!']);
});

Route::get('/health', function () {
    return response()->json(['status' => 'healthy']);
});

Route::get('/users', function () {
    return response()->json([
        'users' => [
            ['id' => 1, 'name' => 'Alice'],
            ['id' => 2, 'name' => 'Bob'],
        ],
    ]);
});

Route::get('/users/{id}', function (int $id) {
    return response()->json(['id' => $id, 'name' => 'Alice']);
});

Route::post('/users', function (Request $request) {
    $data = $request->validate([
        'name' => 'required|string|max:255',
    ]);

    return response()->json(['id' => 3, 'name' => $data['name']], 201);
});
