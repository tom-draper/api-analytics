<?php

declare(strict_types=1);

namespace ApiAnalytics\PHP;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use ApiAnalytics\Core\RequestData;

/**
 * Analytics client for native PHP applications.
 *
 * Usage:
 *   $analytics = new Analytics('your-api-key');
 *   $startTime = microtime(true);
 *   // ... your API logic ...
 *   $responseTimeMs = (int) round((microtime(true) - $startTime) * 1000);
 *   $analytics->log($_SERVER, $responseTimeMs, $statusCode);
 */
class Analytics
{
    private Client $client;

    public function __construct(string $apiKey, ?Config $config = null)
    {
        $this->client = new Client($apiKey, 'Native PHP', $config);
    }

    /**
     * Log a request using the standard $_SERVER superglobal.
     *
     * @param array $server The $_SERVER superglobal or equivalent
     * @param int $responseTime Response time in milliseconds
     * @param int $status HTTP status code
     */
    public function log(array $server, int $responseTime, int $status): void
    {
        $requestData = $this->client->createRequestData($server, $responseTime, $status);
        $this->client->logRequest($requestData);
    }

    /**
     * Log a pre-built RequestData object.
     */
    public function logRequest(RequestData $requestData): void
    {
        $this->client->logRequest($requestData);
    }

    /**
     * Force flush all buffered requests.
     */
    public function flush(): void
    {
        $this->client->flush();
    }

    /**
     * Get the underlying client.
     */
    public function getClient(): Client
    {
        return $this->client;
    }

    /**
     * Get the configuration.
     */
    public function getConfig(): Config
    {
        return $this->client->getConfig();
    }
}
