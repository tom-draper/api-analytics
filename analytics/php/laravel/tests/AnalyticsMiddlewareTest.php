<?php

declare(strict_types=1);

namespace ApiAnalytics\Laravel\Tests;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use ApiAnalytics\Laravel\AnalyticsMiddleware;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use PHPUnit\Framework\TestCase;

class AnalyticsMiddlewareTest extends TestCase
{
    private function createMockClient(): Client
    {
        $config = new Config();
        $config->serverUrl = 'http://localhost:99999/'; // Prevent actual requests

        return new Client('test-api-key', 'Laravel', $config);
    }

    public function testMiddlewareLogsRequest(): void
    {
        $client = $this->createMockClient();
        $middleware = new AnalyticsMiddleware($client);

        $request = Request::create('/api/users', 'GET', [], [], [], [
            'HTTP_HOST' => 'api.example.com',
            'HTTP_USER_AGENT' => 'Mozilla/5.0',
            'REMOTE_ADDR' => '192.168.1.1',
        ]);

        $response = new Response('OK', 200);

        $result = $middleware->handle($request, function ($req) use ($response) {
            return $response;
        });

        $this->assertSame(200, $result->getStatusCode());
        $this->assertSame(1, $client->getBufferSize());
    }

    public function testMiddlewarePreservesResponseStatus(): void
    {
        $client = $this->createMockClient();
        $middleware = new AnalyticsMiddleware($client);

        $request = Request::create('/api/users', 'POST');

        $result = $middleware->handle($request, function ($req) {
            return new Response('Created', 201);
        });

        $this->assertSame(201, $result->getStatusCode());
    }

    public function testMiddlewareHandles404(): void
    {
        $client = $this->createMockClient();
        $middleware = new AnalyticsMiddleware($client);

        $request = Request::create('/not-found', 'GET');

        $result = $middleware->handle($request, function ($req) {
            return new Response('Not Found', 404);
        });

        $this->assertSame(404, $result->getStatusCode());
        $this->assertSame(1, $client->getBufferSize());
    }

    public function testMiddlewareHandles500(): void
    {
        $client = $this->createMockClient();
        $middleware = new AnalyticsMiddleware($client);

        $request = Request::create('/api/error', 'GET');

        $result = $middleware->handle($request, function ($req) {
            return new Response('Internal Server Error', 500);
        });

        $this->assertSame(500, $result->getStatusCode());
        $this->assertSame(1, $client->getBufferSize());
    }

    public function testMiddlewareWithDifferentMethods(): void
    {
        $client = $this->createMockClient();
        $middleware = new AnalyticsMiddleware($client);

        $methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

        foreach ($methods as $method) {
            $request = Request::create('/api/test', $method);

            $middleware->handle($request, function ($req) {
                return new Response('OK', 200);
            });
        }

        $this->assertSame(5, $client->getBufferSize());
    }
}
