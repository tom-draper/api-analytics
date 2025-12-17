<?php

declare(strict_types=1);

namespace ApiAnalytics\Core;

/**
 * Data transfer object representing a single API request to be logged.
 */
class RequestData
{
    public string $hostname;
    public ?string $ipAddress;
    public string $userAgent;
    public string $path;
    public string $method;
    public int $responseTime;
    public int $status;
    public ?string $userId;
    public string $createdAt;

    public function __construct(
        string $hostname,
        ?string $ipAddress,
        string $userAgent,
        string $path,
        string $method,
        int $responseTime,
        int $status,
        ?string $userId,
        ?string $createdAt = null
    ) {
        $this->hostname = $hostname;
        $this->ipAddress = $ipAddress;
        $this->userAgent = $userAgent;
        $this->path = $path;
        $this->method = $method;
        $this->responseTime = $responseTime;
        $this->status = $status;
        $this->userId = $userId;
        $this->createdAt = $createdAt ?? gmdate('Y-m-d\TH:i:s\Z');
    }

    /**
     * Convert to array format expected by the API.
     *
     * @return array<string, mixed>
     */
    public function toArray(): array
    {
        return [
            'hostname' => $this->hostname,
            'ip_address' => $this->ipAddress,
            'user_agent' => $this->userAgent,
            'path' => $this->path,
            'method' => $this->method,
            'response_time' => $this->responseTime,
            'status' => $this->status,
            'user_id' => $this->userId,
            'created_at' => $this->createdAt,
        ];
    }
}
