<?php

declare(strict_types=1);

namespace ApiAnalytics\Core;

/**
 * Configuration for API Analytics middleware.
 *
 * Allows customization of data extraction and privacy settings.
 */
class Config
{
    /**
     * Controls client identification by IP address.
     *
     * - 0: Sends client IP to the server to be stored and client location is inferred.
     * - 1: Sends the client IP to the server only for the location to be inferred
     *      and stored, with the IP discarded afterwards.
     * - 2: Avoids sending the client IP address to the server.
     */
    public int $privacyLevel = 0;

    /**
     * Server URL for self-hosting.
     */
    public string $serverUrl = 'https://www.apianalytics-server.com/';

    /** @var callable(array): string */
    private $getPath;

    /** @var callable(array): string */
    private $getHostname;

    /** @var callable(array): ?string */
    private $getIpAddress;

    /** @var callable(array): string */
    private $getUserAgent;

    /** @var callable(array): ?string */
    private $getUserId;

    public function __construct()
    {
        $this->getPath = [Mappers::class, 'getPath'];
        $this->getHostname = [Mappers::class, 'getHostname'];
        $this->getIpAddress = [Mappers::class, 'getIpAddress'];
        $this->getUserAgent = [Mappers::class, 'getUserAgent'];
        $this->getUserId = [Mappers::class, 'getUserId'];
    }

    public function setGetPath(callable $callback): self
    {
        $this->getPath = $callback;
        return $this;
    }

    public function setGetHostname(callable $callback): self
    {
        $this->getHostname = $callback;
        return $this;
    }

    public function setGetIpAddress(callable $callback): self
    {
        $this->getIpAddress = $callback;
        return $this;
    }

    public function setGetUserAgent(callable $callback): self
    {
        $this->getUserAgent = $callback;
        return $this;
    }

    public function setGetUserId(callable $callback): self
    {
        $this->getUserId = $callback;
        return $this;
    }

    public function extractPath(array $context): string
    {
        return ($this->getPath)($context);
    }

    public function extractHostname(array $context): string
    {
        return ($this->getHostname)($context);
    }

    public function extractIpAddress(array $context): ?string
    {
        if ($this->privacyLevel >= 2) {
            return null;
        }
        return ($this->getIpAddress)($context);
    }

    public function extractUserAgent(array $context): string
    {
        return ($this->getUserAgent)($context);
    }

    public function extractUserId(array $context): ?string
    {
        return ($this->getUserId)($context);
    }

    public function getApiEndpoint(): string
    {
        return rtrim($this->serverUrl, '/') . '/api/log-request';
    }
}
