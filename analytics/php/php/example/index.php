<?php

/**
 * Native PHP Example for API Analytics
 *
 * This example demonstrates how to integrate API Analytics
 * into a simple native PHP application.
 */

declare(strict_types=1);

require_once __DIR__ . '/../vendor/autoload.php';

use ApiAnalytics\Core\Config;
use ApiAnalytics\PHP\Analytics;

// Configuration
$apiKey = getenv('API_ANALYTICS_KEY') ?: 'your-api-key-here';

// Optional: Custom configuration
$config = new Config();
$config->privacyLevel = 0;

// Optional: Custom user ID extraction
$config->setGetUserId(fn(array $ctx) => $ctx['HTTP_X_API_KEY'] ?? null);

// Initialize analytics
$analytics = new Analytics($apiKey, $config);

// Start timing
$startTime = microtime(true);

// ============================================
// Your API Logic Here
// ============================================

$requestUri = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH);
$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';

$response = null;
$statusCode = 200;

switch ($requestUri) {
    case '/':
    case '/api':
        $response = ['message' => 'Hello World!', 'timestamp' => time()];
        break;

    case '/api/health':
        $response = ['status' => 'healthy'];
        break;

    case '/api/users':
        if ($method === 'GET') {
            $response = ['users' => [['id' => 1, 'name' => 'Alice']]];
        } else {
            $response = ['error' => 'Method not allowed'];
            $statusCode = 405;
        }
        break;

    default:
        $response = ['error' => 'Not found'];
        $statusCode = 404;
}

// ============================================
// End of API Logic
// ============================================

// Calculate response time in milliseconds
$responseTimeMs = (int) round((microtime(true) - $startTime) * 1000);

// Log the request
$analytics->log($_SERVER, $responseTimeMs, $statusCode);

// Send response
header('Content-Type: application/json');
http_response_code($statusCode);
echo json_encode($response);
