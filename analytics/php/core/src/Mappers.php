<?php

declare(strict_types=1);

namespace ApiAnalytics\Core;

/**
 * Default mapper functions for extracting request data.
 *
 * These functions provide sensible defaults and can be overridden via Config.
 */
class Mappers
{
    /**
     * Extract the request path from the context.
     */
    public static function getPath(array $context): string
    {
        if (isset($context['REQUEST_URI'])) {
            $path = parse_url($context['REQUEST_URI'], PHP_URL_PATH);
            return $path ?: '/';
        }
        return '/';
    }

    /**
     * Extract the hostname from the context.
     */
    public static function getHostname(array $context): string
    {
        return $context['HTTP_HOST'] ?? $context['SERVER_NAME'] ?? '';
    }

    /**
     * Extract the client IP address from the context.
     *
     * Checks common proxy headers in order of preference:
     * 1. CF-Connecting-IP (Cloudflare)
     * 2. X-Forwarded-For (standard proxy header, uses first IP)
     * 3. X-Real-IP (nginx)
     * 4. REMOTE_ADDR (direct connection)
     */
    public static function getIpAddress(array $context): ?string
    {
        // Cloudflare
        if (isset($context['HTTP_CF_CONNECTING_IP'])) {
            return $context['HTTP_CF_CONNECTING_IP'];
        }

        // Standard proxy header (proxy/load balancer)
        $forwardedFor = $context['HTTP_X_FORWARDED_FOR'] ?? null;
        if ($forwardedFor !== null) {
            $ips = array_map('trim', explode(',', $forwardedFor));
            return $ips[0] ?? null;
        }

        // Nginx real IP header
        if (isset($context['HTTP_X_REAL_IP'])) {
            return $context['HTTP_X_REAL_IP'];
        }

        return $context['REMOTE_ADDR'] ?? null;
    }

    /**
     * Extract the user agent from the context.
     */
    public static function getUserAgent(array $context): string
    {
        return $context['HTTP_USER_AGENT'] ?? '';
    }

    /**
     * Extract a custom user ID from the context.
     * Override via Config to provide custom user identification.
     */
    public static function getUserId(array $context): ?string
    {
        return null;
    }

    /**
     * Extract the HTTP method from the context.
     */
    public static function getMethod(array $context): string
    {
        return $context['REQUEST_METHOD'] ?? 'GET';
    }
}
