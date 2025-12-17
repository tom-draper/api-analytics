<?php

declare(strict_types=1);

namespace ApiAnalytics\Core\Tests;

use ApiAnalytics\Core\Mappers;
use PHPUnit\Framework\TestCase;

class MappersTest extends TestCase
{
    public function testGetPath(): void
    {
        $this->assertSame('/api/users', Mappers::getPath(['REQUEST_URI' => '/api/users']));
        $this->assertSame('/api/users', Mappers::getPath(['REQUEST_URI' => '/api/users?page=1']));
        $this->assertSame('/', Mappers::getPath(['REQUEST_URI' => '/']));
        $this->assertSame('/', Mappers::getPath([]));
    }

    public function testGetHostname(): void
    {
        $this->assertSame('api.example.com', Mappers::getHostname(['HTTP_HOST' => 'api.example.com']));
        $this->assertSame('server.example.com', Mappers::getHostname(['SERVER_NAME' => 'server.example.com']));
        $this->assertSame('api.example.com', Mappers::getHostname([
            'HTTP_HOST' => 'api.example.com',
            'SERVER_NAME' => 'server.example.com'
        ]));
        $this->assertSame('', Mappers::getHostname([]));
    }

    public function testGetIpAddress(): void
    {
        // Direct REMOTE_ADDR
        $this->assertSame('192.168.1.1', Mappers::getIpAddress(['REMOTE_ADDR' => '192.168.1.1']));

        // CF-Connecting-IP takes highest precedence (Cloudflare)
        $this->assertSame('203.0.113.1', Mappers::getIpAddress([
            'HTTP_CF_CONNECTING_IP' => '203.0.113.1',
            'HTTP_X_FORWARDED_FOR' => '10.0.0.1',
            'HTTP_X_REAL_IP' => '10.0.0.5',
            'REMOTE_ADDR' => '192.168.1.1'
        ]));

        // X-Forwarded-For takes precedence over X-Real-IP and REMOTE_ADDR
        $this->assertSame('10.0.0.1', Mappers::getIpAddress([
            'HTTP_X_FORWARDED_FOR' => '10.0.0.1',
            'REMOTE_ADDR' => '192.168.1.1'
        ]));

        // X-Forwarded-For with multiple IPs (uses first)
        $this->assertSame('10.0.0.1', Mappers::getIpAddress([
            'HTTP_X_FORWARDED_FOR' => '10.0.0.1, 10.0.0.2, 10.0.0.3'
        ]));

        // X-Real-IP (nginx) takes precedence over REMOTE_ADDR
        $this->assertSame('10.0.0.5', Mappers::getIpAddress([
            'HTTP_X_REAL_IP' => '10.0.0.5',
            'REMOTE_ADDR' => '192.168.1.1'
        ]));

        // Empty context
        $this->assertNull(Mappers::getIpAddress([]));
    }

    public function testGetUserAgent(): void
    {
        $this->assertSame('Mozilla/5.0', Mappers::getUserAgent(['HTTP_USER_AGENT' => 'Mozilla/5.0']));
        $this->assertSame('', Mappers::getUserAgent([]));
    }

    public function testGetUserId(): void
    {
        // Default always returns null
        $this->assertNull(Mappers::getUserId([]));
        $this->assertNull(Mappers::getUserId(['HTTP_X_API_KEY' => 'some-key']));
    }

    public function testGetMethod(): void
    {
        $this->assertSame('GET', Mappers::getMethod(['REQUEST_METHOD' => 'GET']));
        $this->assertSame('POST', Mappers::getMethod(['REQUEST_METHOD' => 'POST']));
        $this->assertSame('PUT', Mappers::getMethod(['REQUEST_METHOD' => 'PUT']));
        $this->assertSame('DELETE', Mappers::getMethod(['REQUEST_METHOD' => 'DELETE']));
        $this->assertSame('PATCH', Mappers::getMethod(['REQUEST_METHOD' => 'PATCH']));
        $this->assertSame('GET', Mappers::getMethod([])); // Default
    }
}
