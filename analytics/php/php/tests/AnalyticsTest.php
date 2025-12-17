<?php

declare(strict_types=1);

namespace ApiAnalytics\PHP\Tests;

use ApiAnalytics\Core\Config;
use ApiAnalytics\Core\RequestData;
use ApiAnalytics\PHP\Analytics;
use PHPUnit\Framework\TestCase;

class AnalyticsTest extends TestCase
{
    public function testConstructor(): void
    {
        $analytics = new Analytics('test-api-key');

        $this->assertSame('Native PHP', $analytics->getClient()->getFramework());
        $this->assertInstanceOf(Config::class, $analytics->getConfig());
    }

    public function testConstructorWithConfig(): void
    {
        $config = new Config();
        $config->privacyLevel = 2;

        $analytics = new Analytics('test-api-key', $config);

        $this->assertSame(2, $analytics->getConfig()->privacyLevel);
    }

    public function testLog(): void
    {
        $config = new Config();
        $config->serverUrl = 'http://localhost:99999/'; // Prevent actual requests

        $analytics = new Analytics('test-api-key', $config);

        $server = [
            'HTTP_HOST' => 'api.example.com',
            'REMOTE_ADDR' => '192.168.1.1',
            'HTTP_USER_AGENT' => 'Mozilla/5.0',
            'REQUEST_URI' => '/api/users',
            'REQUEST_METHOD' => 'GET',
        ];

        $analytics->log($server, 42, 200);

        $this->assertSame(1, $analytics->getClient()->getBufferSize());
    }

    public function testLogRequest(): void
    {
        $config = new Config();
        $config->serverUrl = 'http://localhost:99999/';

        $analytics = new Analytics('test-api-key', $config);

        $requestData = new RequestData(
            'hostname',
            '1.2.3.4',
            'ua',
            '/path',
            'POST',
            100,
            201,
            null
        );

        $analytics->logRequest($requestData);

        $this->assertSame(1, $analytics->getClient()->getBufferSize());
    }

    public function testFlush(): void
    {
        $config = new Config();
        $config->serverUrl = 'http://localhost:99999/';

        $analytics = new Analytics('test-api-key', $config);

        $analytics->log($_SERVER + [
            'HTTP_HOST' => 'test',
            'REQUEST_URI' => '/',
            'REQUEST_METHOD' => 'GET',
        ], 50, 200);

        $this->assertSame(1, $analytics->getClient()->getBufferSize());

        $analytics->flush();

        $this->assertSame(0, $analytics->getClient()->getBufferSize());
    }
}
