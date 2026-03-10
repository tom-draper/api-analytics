<?php

declare(strict_types=1);

namespace ApiAnalytics\Symfony;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Event\RequestEvent;
use Symfony\Component\HttpKernel\Event\ResponseEvent;
use Symfony\Component\HttpKernel\KernelEvents;

/**
 * Symfony event subscriber for API Analytics.
 *
 * Automatically logs all incoming requests to the API Analytics service.
 *
 * Register as a service in services.yaml:
 *
 *     ApiAnalytics\Symfony\AnalyticsEventSubscriber:
 *         arguments:
 *             $apiKey: '%env(API_ANALYTICS_KEY)%'
 *         tags:
 *             - { name: kernel.event_subscriber }
 */
class AnalyticsEventSubscriber implements EventSubscriberInterface
{
    private Client $client;

    public function __construct(string $apiKey, ?string $serverUrl = null, int $privacyLevel = 0)
    {
        $config = new Config();
        $config->privacyLevel = $privacyLevel;

        if ($serverUrl !== null) {
            $config->serverUrl = $serverUrl;
        }

        $this->client = new Client($apiKey, 'Symfony', $config);
    }

    public static function getSubscribedEvents(): array
    {
        return [
            KernelEvents::REQUEST => ['onKernelRequest', 256],
            KernelEvents::RESPONSE => ['onKernelResponse', -256],
        ];
    }

    public function onKernelRequest(RequestEvent $event): void
    {
        if (!$event->isMainRequest()) {
            return;
        }

        $event->getRequest()->attributes->set('_analytics_start', microtime(true));
    }

    public function onKernelResponse(ResponseEvent $event): void
    {
        if (!$event->isMainRequest()) {
            return;
        }

        $request = $event->getRequest();
        $response = $event->getResponse();

        $startTime = $request->attributes->get('_analytics_start');
        if ($startTime === null) {
            return;
        }

        $responseTimeMs = (int) round((microtime(true) - $startTime) * 1000);

        $context = $this->buildContext($request);

        $requestData = $this->client->createRequestData(
            $context,
            $responseTimeMs,
            $response->getStatusCode()
        );

        $this->client->logRequest($requestData);
    }

    /**
     * Build context array from Symfony request.
     *
     * @return array<string, mixed>
     */
    private function buildContext(Request $request): array
    {
        return [
            'REQUEST_METHOD' => $request->getMethod(),
            'REQUEST_URI' => $request->getRequestUri(),
            'HTTP_HOST' => $request->getHost(),
            'HTTP_USER_AGENT' => $request->headers->get('User-Agent', ''),
            'REMOTE_ADDR' => $request->getClientIp(),
            'HTTP_X_FORWARDED_FOR' => $request->headers->get('X-Forwarded-For'),
            'HTTP_X_REAL_IP' => $request->headers->get('X-Real-IP'),
            'HTTP_CF_CONNECTING_IP' => $request->headers->get('CF-Connecting-IP'),
        ];
    }

    public function getClient(): Client
    {
        return $this->client;
    }
}
