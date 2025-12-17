<?php

declare(strict_types=1);

namespace ApiAnalytics\Core\Tests;

use ApiAnalytics\Core\RequestData;
use PHPUnit\Framework\TestCase;

class RequestDataTest extends TestCase
{
    public function testConstructor(): void
    {
        $requestData = new RequestData(
            'api.example.com',
            '192.168.1.1',
            'Mozilla/5.0',
            '/api/users',
            'GET',
            42,
            200,
            'user-123',
            '2024-01-15T10:30:00Z'
        );

        $this->assertSame('api.example.com', $requestData->hostname);
        $this->assertSame('192.168.1.1', $requestData->ipAddress);
        $this->assertSame('Mozilla/5.0', $requestData->userAgent);
        $this->assertSame('/api/users', $requestData->path);
        $this->assertSame('GET', $requestData->method);
        $this->assertSame(42, $requestData->responseTime);
        $this->assertSame(200, $requestData->status);
        $this->assertSame('user-123', $requestData->userId);
        $this->assertSame('2024-01-15T10:30:00Z', $requestData->createdAt);
    }

    public function testConstructorWithNullValues(): void
    {
        $requestData = new RequestData(
            'api.example.com',
            null,
            'Mozilla/5.0',
            '/api/users',
            'POST',
            100,
            201,
            null
        );

        $this->assertNull($requestData->ipAddress);
        $this->assertNull($requestData->userId);
        $this->assertNotEmpty($requestData->createdAt);
    }

    public function testDefaultCreatedAt(): void
    {
        $before = gmdate('Y-m-d\TH:i:s\Z');

        $requestData = new RequestData(
            'hostname',
            null,
            'ua',
            '/path',
            'GET',
            0,
            200,
            null
        );

        $after = gmdate('Y-m-d\TH:i:s\Z');

        // createdAt should be between before and after (or equal to either)
        $this->assertGreaterThanOrEqual($before, $requestData->createdAt);
        $this->assertLessThanOrEqual($after, $requestData->createdAt);
    }

    public function testToArray(): void
    {
        $requestData = new RequestData(
            'api.example.com',
            '192.168.1.1',
            'Mozilla/5.0',
            '/api/users',
            'GET',
            42,
            200,
            'user-123',
            '2024-01-15T10:30:00Z'
        );

        $expected = [
            'hostname' => 'api.example.com',
            'ip_address' => '192.168.1.1',
            'user_agent' => 'Mozilla/5.0',
            'path' => '/api/users',
            'method' => 'GET',
            'response_time' => 42,
            'status' => 200,
            'user_id' => 'user-123',
            'created_at' => '2024-01-15T10:30:00Z',
        ];

        $this->assertSame($expected, $requestData->toArray());
    }

    public function testToArrayWithNullValues(): void
    {
        $requestData = new RequestData(
            'hostname',
            null,
            'ua',
            '/path',
            'DELETE',
            50,
            204,
            null,
            '2024-01-15T10:30:00Z'
        );

        $array = $requestData->toArray();

        $this->assertNull($array['ip_address']);
        $this->assertNull($array['user_id']);
        $this->assertSame('DELETE', $array['method']);
        $this->assertSame(204, $array['status']);
    }
}
