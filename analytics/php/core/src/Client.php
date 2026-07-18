<?php

declare(strict_types=1);

namespace ApiAnalytics\Core;

/**
 * Core analytics client that handles request buffering and batched posting.
 *
 * Requests are buffered and sent to the API Analytics server every 60 seconds
 * or when the script terminates (via shutdown function).
 */
class Client
{
    private const FLUSH_INTERVAL_SECONDS = 60;

    // Analytics posting is best-effort, so keep timeouts short: the flush runs
    // synchronously at request shutdown, and a slow or unreachable server must
    // not tie up the PHP-FPM worker for long.
    private const POST_TIMEOUT_SECONDS = 5;
    private const CONNECT_TIMEOUT_SECONDS = 2;

    private string $apiKey;
    private string $framework;
    private Config $config;

    /** @var RequestData[] */
    private array $requests = [];
    private float $lastPosted;
    private bool $shutdownRegistered = false;

    public function __construct(string $apiKey, string $framework, ?Config $config = null)
    {
        $this->apiKey = $apiKey;
        $this->framework = $framework;
        $this->config = $config ?? new Config();
        $this->lastPosted = microtime(true);

        $this->registerShutdown();
    }

    /**
     * Log a request to the buffer.
     */
    public function logRequest(RequestData $requestData): void
    {
        if (empty($this->apiKey)) {
            return;
        }

        $this->requests[] = $requestData;

        $now = microtime(true);
        if (($now - $this->lastPosted) > self::FLUSH_INTERVAL_SECONDS && count($this->requests) > 0) {
            $this->postRequests();
        }
    }

    /**
     * Create RequestData from a request context.
     */
    public function createRequestData(array $context, int $responseTime, int $status): RequestData
    {
        return new RequestData(
            $this->config->extractHostname($context),
            $this->config->extractIpAddress($context),
            $this->config->extractUserAgent($context),
            $this->config->extractPath($context),
            Mappers::getMethod($context),
            $responseTime,
            $status,
            $this->config->extractUserId($context)
        );
    }

    /**
     * Force flush all buffered requests to the server.
     */
    public function flush(): void
    {
        if (count($this->requests) > 0) {
            $this->postRequests();
        }
    }

    /**
     * Get the current buffer size.
     */
    public function getBufferSize(): int
    {
        return count($this->requests);
    }

    /**
     * Get the configuration instance.
     */
    public function getConfig(): Config
    {
        return $this->config;
    }

    /**
     * Get the framework name.
     */
    public function getFramework(): string
    {
        return $this->framework;
    }

    private function registerShutdown(): void
    {
        if (!$this->shutdownRegistered) {
            register_shutdown_function([$this, 'shutdownFlush']);
            $this->shutdownRegistered = true;
        }
    }

    /**
     * Flush handler run at script shutdown. On PHP-FPM it finalises the HTTP
     * response first, so the buffered analytics POST runs after the client has
     * already received the response and cannot add to its latency.
     */
    public function shutdownFlush(): void
    {
        if (function_exists('fastcgi_finish_request')) {
            fastcgi_finish_request();
        }

        $this->flush();
    }

    private function postRequests(): void
    {
        if (empty($this->apiKey) || count($this->requests) === 0) {
            return;
        }

        $payload = [
            'api_key' => $this->apiKey,
            'requests' => array_map(fn(RequestData $r) => $r->toArray(), $this->requests),
            'framework' => $this->framework,
            'privacy_level' => $this->config->privacyLevel,
        ];

        $this->sendRequest($payload);
        $this->requests = [];
        $this->lastPosted = microtime(true);
    }

    private function sendRequest(array $payload): void
    {
        $url = $this->config->getApiEndpoint();
        $json = json_encode($payload);

        if ($json === false) {
            return;
        }

        if (function_exists('curl_init')) {
            $this->sendWithCurl($url, $json);
            return;
        }

        $this->sendWithFileGetContents($url, $json);
    }

    private function sendWithCurl(string $url, string $json): void
    {
        $ch = curl_init($url);
        if ($ch === false) {
            return;
        }

        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => $json,
            CURLOPT_HTTPHEADER => [
                'Content-Type: application/json',
                'Content-Length: ' . strlen($json),
            ],
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => self::POST_TIMEOUT_SECONDS,
            CURLOPT_CONNECTTIMEOUT => self::CONNECT_TIMEOUT_SECONDS,
        ]);

        curl_exec($ch);
        curl_close($ch);
    }

    private function sendWithFileGetContents(string $url, string $json): void
    {
        $context = stream_context_create([
            'http' => [
                'method' => 'POST',
                'header' => "Content-Type: application/json\r\nContent-Length: " . strlen($json) . "\r\n",
                'content' => $json,
                'timeout' => self::POST_TIMEOUT_SECONDS,
                'ignore_errors' => true,
            ],
        ]);

        @file_get_contents($url, false, $context);
    }
}
