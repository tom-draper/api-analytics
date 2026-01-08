# API Analytics for PHP

A free and lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to [apianalytics.dev/generate](https://apianalytics.dev/generate) to generate your unique API key with a single click. This key is used to monitor your specific API and should be stored privately.

### 2. Install the package

```bash
composer require api-analytics/php
```

### 3. Add analytics to your API

```php
<?php

require_once 'vendor/autoload.php';

use ApiAnalytics\Core\Config;
use ApiAnalytics\PHP\Analytics;

$analytics = new Analytics('your-api-key');

// Start timing
$startTime = microtime(true);

// Your API logic here
$response = ['message' => 'Hello World!'];
$statusCode = 200;

// Calculate response time in milliseconds
$responseTimeMs = (int) round((microtime(true) - $startTime) * 1000);

// Log the request
$analytics->log($_SERVER, $responseTimeMs, $statusCode);

// Send response
header('Content-Type: application/json');
http_response_code($statusCode);
echo json_encode($response);
```

### 4. View your analytics

Your API will now log and store incoming request data on all routes. View your analytics at [apianalytics.dev/dashboard](https://apianalytics.dev/dashboard).

## Customisation

Custom mapping functions can override the default behaviour:

```php
use ApiAnalytics\Core\Config;
use ApiAnalytics\PHP\Analytics;

$config = new Config();

// Custom IP extraction
$config->setGetIpAddress(function (array $ctx) {
    return $ctx['HTTP_X_REAL_IP'] ?? $ctx['REMOTE_ADDR'] ?? null;
});

// Custom user ID from API key header
$config->setGetUserId(function (array $ctx) {
    return $ctx['HTTP_X_API_KEY'] ?? null;
});

$analytics = new Analytics('your-api-key', $config);
```

## Privacy Levels

Control IP address handling:

- `0` - IP stored, location inferred (default)
- `1` - Location inferred, IP discarded
- `2` - IP never sent

```php
$config = new Config();
$config->privacyLevel = 2;
```

## Self-Hosting

For self-hosted instances:

```php
$config = new Config();
$config->serverUrl = 'https://your-server.com/';

$analytics = new Analytics('your-api-key', $config);
```

## More Information

- [Dashboard](https://apianalytics.dev/dashboard)
- [Documentation](https://github.com/tom-draper/api-analytics)
- [Privacy Policy](https://www.apianalytics.dev/privacy-policy)
