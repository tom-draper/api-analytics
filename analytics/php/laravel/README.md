# API Analytics for Laravel

A free and lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to [apianalytics.dev/generate](https://apianalytics.dev/generate) to generate your unique API key with a single click.

### 2. Install the package

```bash
composer require api-analytics/laravel
```

### 3. Add your API key

Add to your `.env` file:

```env
API_ANALYTICS_KEY=your-api-key-here
```

### 4. Register the middleware

Add to `app/Http/Kernel.php`:

```php
// Global middleware (all routes)
protected $middleware = [
    // ... other middleware
    \ApiAnalytics\Laravel\AnalyticsMiddleware::class,
];

// Or for API routes only
protected $middlewareGroups = [
    'api' => [
        // ... other middleware
        \ApiAnalytics\Laravel\AnalyticsMiddleware::class,
    ],
];
```

### 5. View your analytics

Your API will now log incoming requests. View your dashboard at [apianalytics.dev/dashboard](https://apianalytics.dev/dashboard).

## Configuration

Publish the config file (optional):

```bash
php artisan vendor:publish --tag=api-analytics-config
```

This creates `config/api-analytics.php`:

```php
return [
    'api_key' => env('API_ANALYTICS_KEY', ''),
    'privacy_level' => env('API_ANALYTICS_PRIVACY_LEVEL', 0),
    'server_url' => env('API_ANALYTICS_SERVER_URL', 'https://www.apianalytics-server.com/'),
    'get_user_id' => null, // Optional callback
];
```

## Custom User ID

Track users by authenticated user or API key:

```php
// In config/api-analytics.php
'get_user_id' => function (array $context) {
    $request = $context['request'] ?? null;

    // Use authenticated user ID
    if ($request?->user()) {
        return (string) $request->user()->id;
    }

    // Or fall back to API key header
    return $request?->header('X-API-Key') ?? '';
},
```

Or in a service provider for more control:

```php
// app/Providers/AppServiceProvider.php
use ApiAnalytics\Core\Config;

public function register(): void
{
    $this->app->singleton(Config::class, function () {
        $config = new Config();
        $config->privacyLevel = 2;

        $config->setGetUserId(function (array $context) {
            $request = $context['request'] ?? null;
            return $request?->user()?->id ?? $request?->header('X-API-Key');
        });

        return $config;
    });
}
```

## Privacy Levels

Control IP address handling via environment:

```env
API_ANALYTICS_PRIVACY_LEVEL=2
```

- `0` - IP stored, location inferred (default)
- `1` - Location inferred, IP discarded
- `2` - IP never sent

## Facade Usage

Access the analytics client directly:

```php
use ApiAnalytics\Laravel\Facade as Analytics;

// Get buffer size
$count = Analytics::getBufferSize();

// Force flush
Analytics::flush();
```

## Testing

Mock the client in tests:

```php
use ApiAnalytics\Core\Client;

$this->mock(Client::class, function ($mock) {
    $mock->shouldReceive('logRequest')->andReturnNull();
    $mock->shouldReceive('createRequestData')->andReturn(
        new \ApiAnalytics\Core\RequestData('', null, '', '/', 'GET', 0, 200, null)
    );
});
```

## Self-Hosting

For self-hosted instances:

```env
API_ANALYTICS_SERVER_URL=https://your-server.com/
```

## More Information

- [Dashboard](https://apianalytics.dev/dashboard)
- [Documentation](https://github.com/tom-draper/api-analytics)
- [Privacy Policy](https://www.apianalytics.dev/privacy-policy)
