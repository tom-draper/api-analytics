<?php

declare(strict_types=1);

namespace ApiAnalytics\PHP\Tests;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use ApiAnalytics\Core\RequestData;
use ApiAnalytics\PHP\Analytics;
use PHPUnit\Framework\TestCase;
use ReflectionClass;

/**
 * End-to-end tests that verify the complete payload format.
 *
 * These tests capture the payload that would be sent to the API Analytics
 * server and verify it matches the expected format.
 */
class E2ETest extends TestCase
{
    /**
     * Helper to extract the payload from Client before it's sent.
     */
    private function capturePayload(Analytics $analytics): array
    {
        // Use reflection to access the private client and extract payload
        $analyticsReflection = new ReflectionClass($analytics);
        $clientProperty = $analyticsReflection->getProperty('client');
        $clientProperty->setAccessible(true);
        $client = $clientProperty->getValue($analytics);

        $clientReflection = new ReflectionClass($client);

        // Get requests buffer
        $requestsProperty = $clientReflection->getProperty('requests');
        $requestsProperty->setAccessible(true);
        $requests = $requestsProperty->getValue($client);

        // Get other properties
        $apiKeyProperty = $clientReflection->getProperty('apiKey');
        $apiKeyProperty->setAccessible(true);
        $apiKey = $apiKeyProperty->getValue($client);

        $frameworkProperty = $clientReflection->getProperty('framework');
        $frameworkProperty->setAccessible(true);
        $framework = $frameworkProperty->getValue($client);

        $configProperty = $clientReflection->getProperty('config');
        $configProperty->setAccessible(true);
        $config = $configProperty->getValue($client);

        // Build the payload exactly as Client::postRequests would
        return [
            'api_key' => $apiKey,
            'requests' => array_map(fn(RequestData $r) => $r->toArray(), $requests),
            'framework' => $framework,
            'privacy_level' => $config->privacyLevel,
        ];
    }

    public function testFullPayloadFormat(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key-12345', $config);

        // Simulate a request context (like $_SERVER)
        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/v1/users?page=1',
            'REQUEST_METHOD' => 'GET',
            'HTTP_USER_AGENT' => 'Mozilla/5.0 (Test Browser)',
            'REMOTE_ADDR' => '192.168.1.100',
            'HTTP_X_FORWARDED_FOR' => '203.0.113.50, 10.0.0.1',
        ];

        // Log a request (don't flush - just capture)
        $analytics->log($serverContext, 150, 200);

        // Capture the payload that would be sent
        $payload = $this->capturePayload($analytics);

        // Verify top-level structure
        $this->assertArrayHasKey('api_key', $payload);
        $this->assertArrayHasKey('requests', $payload);
        $this->assertArrayHasKey('framework', $payload);
        $this->assertArrayHasKey('privacy_level', $payload);

        $this->assertSame('test-api-key-12345', $payload['api_key']);
        $this->assertSame('Native PHP', $payload['framework']);
        $this->assertSame(0, $payload['privacy_level']);

        // Verify request data structure
        $this->assertCount(1, $payload['requests']);
        $request = $payload['requests'][0];

        // Verify all required fields exist
        $requiredFields = [
            'hostname', 'ip_address', 'user_agent', 'path',
            'method', 'response_time', 'status', 'user_id', 'created_at'
        ];
        foreach ($requiredFields as $field) {
            $this->assertArrayHasKey($field, $request, "Missing field: {$field}");
        }

        // Verify values
        $this->assertSame('api.example.com', $request['hostname']);
        $this->assertSame('203.0.113.50', $request['ip_address']); // First IP from X-Forwarded-For
        $this->assertSame('Mozilla/5.0 (Test Browser)', $request['user_agent']);
        $this->assertSame('/api/v1/users', $request['path']); // Query string stripped
        $this->assertSame('GET', $request['method']);
        $this->assertSame(150, $request['response_time']);
        $this->assertSame(200, $request['status']);
        $this->assertNull($request['user_id']);

        // Verify ISO8601 timestamp format (UTC)
        $this->assertMatchesRegularExpression(
            '/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/',
            $request['created_at'],
            'created_at should be ISO8601 format in UTC'
        );

        // Verify the JSON encoding produces valid JSON
        $json = json_encode($payload);
        $this->assertNotFalse($json, 'Payload should be JSON-encodable');
        $decoded = json_decode($json, true);
        $this->assertEquals($payload, $decoded, 'JSON round-trip should preserve data');
    }

    public function testPrivacyLevel2HidesIp(): void
    {
        $config = new Config();
        $config->privacyLevel = 2;

        $analytics = new Analytics('test-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/users',
            'REQUEST_METHOD' => 'POST',
            'HTTP_USER_AGENT' => 'TestAgent/1.0',
            'REMOTE_ADDR' => '192.168.1.100',
            'HTTP_X_FORWARDED_FOR' => '203.0.113.50',
        ];

        $analytics->log($serverContext, 50, 201);

        $payload = $this->capturePayload($analytics);

        $this->assertSame(2, $payload['privacy_level']);
        $this->assertNull($payload['requests'][0]['ip_address']);
    }

    public function testPrivacyLevel1StillSendsIp(): void
    {
        $config = new Config();
        $config->privacyLevel = 1;

        $analytics = new Analytics('test-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/users',
            'REQUEST_METHOD' => 'GET',
            'HTTP_USER_AGENT' => 'TestAgent/1.0',
            'REMOTE_ADDR' => '192.168.1.100',
        ];

        $analytics->log($serverContext, 50, 200);

        $payload = $this->capturePayload($analytics);

        $this->assertSame(1, $payload['privacy_level']);
        // IP should still be sent at level 1 (server will discard after geolocation)
        $this->assertSame('192.168.1.100', $payload['requests'][0]['ip_address']);
    }

    public function testCloudflareIpHeaderPrecedence(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/users',
            'REQUEST_METHOD' => 'GET',
            'HTTP_USER_AGENT' => 'TestAgent/1.0',
            'HTTP_CF_CONNECTING_IP' => '198.51.100.1', // Cloudflare header (highest priority)
            'HTTP_X_FORWARDED_FOR' => '10.0.0.1',
            'HTTP_X_REAL_IP' => '10.0.0.2',
            'REMOTE_ADDR' => '192.168.1.1',
        ];

        $analytics->log($serverContext, 25, 200);

        $payload = $this->capturePayload($analytics);

        // CF-Connecting-IP should take precedence
        $this->assertSame('198.51.100.1', $payload['requests'][0]['ip_address']);
    }

    public function testXForwardedForPrecedenceOverXRealIp(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/users',
            'REQUEST_METHOD' => 'GET',
            'HTTP_USER_AGENT' => 'TestAgent/1.0',
            'HTTP_X_FORWARDED_FOR' => '10.0.0.1',
            'HTTP_X_REAL_IP' => '10.0.0.2',
            'REMOTE_ADDR' => '192.168.1.1',
        ];

        $analytics->log($serverContext, 25, 200);

        $payload = $this->capturePayload($analytics);

        // X-Forwarded-For should take precedence over X-Real-IP
        $this->assertSame('10.0.0.1', $payload['requests'][0]['ip_address']);
    }

    public function testCustomUserIdMapper(): void
    {
        $config = new Config();
        $config->setGetUserId(function (array $ctx) {
            return $ctx['HTTP_X_API_KEY'] ?? null;
        });

        $analytics = new Analytics('test-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'api.example.com',
            'REQUEST_URI' => '/api/data',
            'REQUEST_METHOD' => 'GET',
            'HTTP_USER_AGENT' => 'TestAgent/1.0',
            'REMOTE_ADDR' => '192.168.1.1',
            'HTTP_X_API_KEY' => 'user-api-key-abc123',
        ];

        $analytics->log($serverContext, 100, 200);

        $payload = $this->capturePayload($analytics);

        $this->assertSame('user-api-key-abc123', $payload['requests'][0]['user_id']);
    }

    public function testBatchedRequests(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key', $config);

        // Log multiple requests
        for ($i = 1; $i <= 3; $i++) {
            $serverContext = [
                'HTTP_HOST' => 'api.example.com',
                'REQUEST_URI' => "/api/endpoint{$i}",
                'REQUEST_METHOD' => 'GET',
                'HTTP_USER_AGENT' => 'TestAgent/1.0',
                'REMOTE_ADDR' => '192.168.1.1',
            ];
            $analytics->log($serverContext, $i * 10, 200);
        }

        $payload = $this->capturePayload($analytics);

        // All 3 requests should be batched together
        $this->assertCount(3, $payload['requests']);
        $this->assertSame('/api/endpoint1', $payload['requests'][0]['path']);
        $this->assertSame('/api/endpoint2', $payload['requests'][1]['path']);
        $this->assertSame('/api/endpoint3', $payload['requests'][2]['path']);

        // Response times should match
        $this->assertSame(10, $payload['requests'][0]['response_time']);
        $this->assertSame(20, $payload['requests'][1]['response_time']);
        $this->assertSame(30, $payload['requests'][2]['response_time']);
    }

    public function testDifferentHttpMethods(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key', $config);

        $methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

        foreach ($methods as $method) {
            $serverContext = [
                'HTTP_HOST' => 'api.example.com',
                'REQUEST_URI' => '/api/resource',
                'REQUEST_METHOD' => $method,
                'HTTP_USER_AGENT' => 'TestAgent/1.0',
                'REMOTE_ADDR' => '192.168.1.1',
            ];
            $analytics->log($serverContext, 50, 200);
        }

        $payload = $this->capturePayload($analytics);

        $this->assertCount(5, $payload['requests']);
        foreach ($methods as $i => $method) {
            $this->assertSame($method, $payload['requests'][$i]['method']);
        }
    }

    public function testDifferentStatusCodes(): void
    {
        $config = new Config();
        $analytics = new Analytics('test-api-key', $config);

        $statusCodes = [200, 201, 204, 400, 401, 403, 404, 500, 502, 503];

        foreach ($statusCodes as $status) {
            $serverContext = [
                'HTTP_HOST' => 'api.example.com',
                'REQUEST_URI' => '/api/resource',
                'REQUEST_METHOD' => 'GET',
                'HTTP_USER_AGENT' => 'TestAgent/1.0',
                'REMOTE_ADDR' => '192.168.1.1',
            ];
            $analytics->log($serverContext, 50, $status);
        }

        $payload = $this->capturePayload($analytics);

        $this->assertCount(count($statusCodes), $payload['requests']);
        foreach ($statusCodes as $i => $status) {
            $this->assertSame($status, $payload['requests'][$i]['status']);
        }
    }

    public function testPayloadMatchesGoFormat(): void
    {
        // This test verifies our payload matches the exact format from Go core.go:
        // type Payload struct {
        //     APIKey       string        `json:"api_key"`
        //     Requests     []RequestData `json:"requests"`
        //     Framework    string        `json:"framework"`
        //     PrivacyLevel int           `json:"privacy_level"`
        // }
        // type RequestData struct {
        //     Hostname     string `json:"hostname"`
        //     IPAddress    string `json:"ip_address"`
        //     Path         string `json:"path"`
        //     UserAgent    string `json:"user_agent"`
        //     Method       string `json:"method"`
        //     ResponseTime int64  `json:"response_time"`
        //     Status       int    `json:"status"`
        //     UserID       string `json:"user_id"`
        //     CreatedAt    string `json:"created_at"`
        // }

        $config = new Config();
        $analytics = new Analytics('my-api-key', $config);

        $serverContext = [
            'HTTP_HOST' => 'example.com',
            'REQUEST_URI' => '/test',
            'REQUEST_METHOD' => 'POST',
            'HTTP_USER_AGENT' => 'Go-http-client/1.1',
            'REMOTE_ADDR' => '127.0.0.1',
        ];

        $analytics->log($serverContext, 100, 200);

        $payload = $this->capturePayload($analytics);

        // Convert to JSON and verify all snake_case keys
        $json = json_encode($payload);
        $this->assertStringContainsString('"api_key":', $json);
        $this->assertStringContainsString('"requests":', $json);
        $this->assertStringContainsString('"framework":', $json);
        $this->assertStringContainsString('"privacy_level":', $json);
        $this->assertStringContainsString('"hostname":', $json);
        $this->assertStringContainsString('"ip_address":', $json);
        $this->assertStringContainsString('"path":', $json);
        $this->assertStringContainsString('"user_agent":', $json);
        $this->assertStringContainsString('"method":', $json);
        $this->assertStringContainsString('"response_time":', $json);
        $this->assertStringContainsString('"status":', $json);
        $this->assertStringContainsString('"user_id":', $json);
        $this->assertStringContainsString('"created_at":', $json);

        // Ensure no camelCase keys leaked through
        $this->assertStringNotContainsString('"apiKey":', $json);
        $this->assertStringNotContainsString('"ipAddress":', $json);
        $this->assertStringNotContainsString('"userAgent":', $json);
        $this->assertStringNotContainsString('"responseTime":', $json);
        $this->assertStringNotContainsString('"userId":', $json);
        $this->assertStringNotContainsString('"createdAt":', $json);
        $this->assertStringNotContainsString('"privacyLevel":', $json);
    }
}
