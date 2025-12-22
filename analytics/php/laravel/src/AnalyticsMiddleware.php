<?php

declare(strict_types=1);

namespace ApiAnalytics\Laravel;

use ApiAnalytics\Core\Client;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Laravel middleware for API Analytics.
 *
 * Automatically logs all incoming requests to the API Analytics service.
 */
class AnalyticsMiddleware
{
    private Client $client;

    public function __construct(Client $client)
    {
        $this->client = $client;
    }

    /**
     * Handle an incoming request.
     *
     * @param Request $request
     * @param Closure(Request): Response $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        $startTime = microtime(true);

        /** @var Response $response */
        $response = $next($request);

        $responseTimeMs = (int) round((microtime(true) - $startTime) * 1000);

        $this->logRequest($request, $response, $responseTimeMs);

        return $response;
    }

    private function logRequest(Request $request, Response $response, int $responseTimeMs): void
    {
        $context = $this->buildContext($request);

        $requestData = $this->client->createRequestData(
            $context,
            $responseTimeMs,
            $response->getStatusCode()
        );

        $this->client->logRequest($requestData);
    }

    /**
     * Build context array from Laravel request.
     *
     * @return array<string, mixed>
     */
    private function buildContext(Request $request): array
    {
        return [
            'request' => $request,
            'REQUEST_METHOD' => $request->getMethod(),
            'REQUEST_URI' => $request->getRequestUri(),
            'HTTP_HOST' => $request->getHost(),
            'HTTP_USER_AGENT' => $request->userAgent() ?? '',
            'REMOTE_ADDR' => $request->ip(),
            'HTTP_X_FORWARDED_FOR' => $request->header('X-Forwarded-For'),
            'HTTP_X_REAL_IP' => $request->header('X-Real-IP'),
        ];
    }
}
