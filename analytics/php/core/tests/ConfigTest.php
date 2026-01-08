<?php

declare(strict_types=1);

namespace ApiAnalytics\Core\Tests;

use ApiAnalytics\Core\Config;
use PHPUnit\Framework\TestCase;

class ConfigTest extends TestCase
{
    public function testDefaultValues(): void
    {
        $config = new Config();

        $this->assertSame(0, $config->privacyLevel);
        $this->assertSame('https://www.apianalytics-server.com/', $config->serverUrl);
    }

    public function testGetApiEndpoint(): void
    {
        $config = new Config();
        $this->assertSame(
            'https://www.apianalytics-server.com/api/log-request',
            $config->getApiEndpoint()
        );

        $config->serverUrl = 'https://custom.server.com';
        $this->assertSame(
            'https://custom.server.com/api/log-request',
            $config->getApiEndpoint()
        );

        $config->serverUrl = 'https://custom.server.com/';
        $this->assertSame(
            'https://custom.server.com/api/log-request',
            $config->getApiEndpoint()
        );
    }

    public function testExtractPath(): void
    {
        $config = new Config();

        $context = ['REQUEST_URI' => '/api/users?page=1'];
        $this->assertSame('/api/users', $config->extractPath($context));

        $context = ['REQUEST_URI' => '/'];
        $this->assertSame('/', $config->extractPath($context));

        $context = [];
        $this->assertSame('/', $config->extractPath($context));
    }

    public function testExtractHostname(): void
    {
        $config = new Config();

        $context = ['HTTP_HOST' => 'api.example.com'];
        $this->assertSame('api.example.com', $config->extractHostname($context));

        $context = ['SERVER_NAME' => 'server.example.com'];
        $this->assertSame('server.example.com', $config->extractHostname($context));

        $context = [];
        $this->assertSame('', $config->extractHostname($context));
    }

    public function testExtractIpAddressWithPrivacyLevel(): void
    {
        $config = new Config();
        $context = ['REMOTE_ADDR' => '192.168.1.1'];

        // Privacy level 0 - IP should be returned
        $config->privacyLevel = 0;
        $this->assertSame('192.168.1.1', $config->extractIpAddress($context));

        // Privacy level 1 - IP should be returned
        $config->privacyLevel = 1;
        $this->assertSame('192.168.1.1', $config->extractIpAddress($context));

        // Privacy level 2 - IP should be null
        $config->privacyLevel = 2;
        $this->assertNull($config->extractIpAddress($context));
    }

    public function testExtractIpAddressFromForwardedHeader(): void
    {
        $config = new Config();

        // X-Forwarded-For with multiple IPs
        $context = ['HTTP_X_FORWARDED_FOR' => '10.0.0.1, 10.0.0.2, 10.0.0.3'];
        $this->assertSame('10.0.0.1', $config->extractIpAddress($context));

        // X-Real-IP
        $context = ['HTTP_X_REAL_IP' => '10.0.0.5'];
        $this->assertSame('10.0.0.5', $config->extractIpAddress($context));
    }

    public function testExtractUserAgent(): void
    {
        $config = new Config();

        $context = ['HTTP_USER_AGENT' => 'Mozilla/5.0'];
        $this->assertSame('Mozilla/5.0', $config->extractUserAgent($context));

        $context = [];
        $this->assertSame('', $config->extractUserAgent($context));
    }

    public function testExtractUserId(): void
    {
        $config = new Config();

        // Default returns null
        $this->assertNull($config->extractUserId([]));

        // Custom mapper
        $config->setGetUserId(fn(array $ctx) => $ctx['HTTP_X_API_KEY'] ?? null);

        $context = ['HTTP_X_API_KEY' => 'user-123'];
        $this->assertSame('user-123', $config->extractUserId($context));
    }

    public function testCustomMappers(): void
    {
        $config = new Config();

        $config->setGetPath(fn(array $ctx) => '/custom/path');
        $config->setGetHostname(fn(array $ctx) => 'custom.host');
        $config->setGetIpAddress(fn(array $ctx) => '1.2.3.4');
        $config->setGetUserAgent(fn(array $ctx) => 'CustomAgent');

        $context = [];
        $this->assertSame('/custom/path', $config->extractPath($context));
        $this->assertSame('custom.host', $config->extractHostname($context));
        $this->assertSame('1.2.3.4', $config->extractIpAddress($context));
        $this->assertSame('CustomAgent', $config->extractUserAgent($context));
    }

    public function testFluentSetters(): void
    {
        $config = new Config();

        $result = $config
            ->setGetPath(fn($ctx) => '/path')
            ->setGetHostname(fn($ctx) => 'host')
            ->setGetIpAddress(fn($ctx) => 'ip')
            ->setGetUserAgent(fn($ctx) => 'ua')
            ->setGetUserId(fn($ctx) => 'user');

        $this->assertSame($config, $result);
    }
}
