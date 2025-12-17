<?php

declare(strict_types=1);

namespace ApiAnalytics\Core\Tests;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use ApiAnalytics\Core\RequestData;
use PHPUnit\Framework\TestCase;

class ClientTest extends TestCase
{
    public function testConstructor(): void
    {
        $client = new Client('test-api-key', 'TestFramework');

        $this->assertSame(0, $client->getBufferSize());
        $this->assertSame('TestFramework', $client->getFramework());
        $this->assertInstanceOf(Config::class, $client->getConfig());
    }

    public function testConstructorWithConfig(): void
    {
        $config = new Config();
        $config->privacyLevel = 2;

        $client = new Client('test-api-key', 'TestFramework', $config);

        $this->assertSame(2, $client->getConfig()->privacyLevel);
    }

    public function testLogRequestWithEmptyApiKey(): void
    {
        $client = new Client('', 'TestFramework');

        $requestData = new RequestData(
            'hostname',
            null,
            'ua',
            '/path',
            'GET',
            100,
            200,
            null
        );

        $client->logRequest($requestData);

        // Buffer should remain empty when API key is empty
        $this->assertSame(0, $client->getBufferSize());
    }

    public function testLogRequestAddsToBuffer(): void
    {
        $client = new Client('test-api-key', 'TestFramework');

        $requestData = new RequestData(
            'hostname',
            '1.2.3.4',
            'ua',
            '/path',
            'GET',
            100,
            200,
            null
        );

        $client->logRequest($requestData);
        $this->assertSame(1, $client->getBufferSize());

        $client->logRequest($requestData);
        $this->assertSame(2, $client->getBufferSize());
    }

    public function testCreateRequestData(): void
    {
        $client = new Client('test-api-key', 'TestFramework');

        $context = [
            'HTTP_HOST' => 'api.example.com',
            'REMOTE_ADDR' => '192.168.1.1',
            'HTTP_USER_AGENT' => 'Mozilla/5.0',
            'REQUEST_URI' => '/api/users?page=1',
            'REQUEST_METHOD' => 'POST',
        ];

        $requestData = $client->createRequestData($context, 42, 201);

        $this->assertSame('api.example.com', $requestData->hostname);
        $this->assertSame('192.168.1.1', $requestData->ipAddress);
        $this->assertSame('Mozilla/5.0', $requestData->userAgent);
        $this->assertSame('/api/users', $requestData->path);
        $this->assertSame('POST', $requestData->method);
        $this->assertSame(42, $requestData->responseTime);
        $this->assertSame(201, $requestData->status);
    }

    public function testCreateRequestDataWithPrivacyLevel2(): void
    {
        $config = new Config();
        $config->privacyLevel = 2;

        $client = new Client('test-api-key', 'TestFramework', $config);

        $context = [
            'HTTP_HOST' => 'api.example.com',
            'REMOTE_ADDR' => '192.168.1.1',
            'REQUEST_URI' => '/path',
            'REQUEST_METHOD' => 'GET',
        ];

        $requestData = $client->createRequestData($context, 100, 200);

        // IP should be null with privacy level 2
        $this->assertNull($requestData->ipAddress);
    }

    public function testFlushClearsBuffer(): void
    {
        // Create a client with a non-existent server URL to prevent actual requests
        $config = new Config();
        $config->serverUrl = 'http://localhost:99999/';

        $client = new Client('test-api-key', 'TestFramework', $config);

        $requestData = new RequestData(
            'hostname',
            null,
            'ua',
            '/path',
            'GET',
            100,
            200,
            null
        );

        $client->logRequest($requestData);
        $client->logRequest($requestData);
        $this->assertSame(2, $client->getBufferSize());

        // Flush will attempt to send (and fail silently), but buffer should be cleared
        $client->flush();
        $this->assertSame(0, $client->getBufferSize());
    }

    public function testGetConfig(): void
    {
        $config = new Config();
        $config->privacyLevel = 1;
        $config->serverUrl = 'https://custom.server.com/';

        $client = new Client('api-key', 'Framework', $config);

        $retrievedConfig = $client->getConfig();
        $this->assertSame(1, $retrievedConfig->privacyLevel);
        $this->assertSame('https://custom.server.com/', $retrievedConfig->serverUrl);
    }
}
