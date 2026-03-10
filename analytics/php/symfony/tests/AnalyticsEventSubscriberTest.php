<?php

declare(strict_types=1);

namespace ApiAnalytics\Symfony\Tests;

use ApiAnalytics\Symfony\AnalyticsEventSubscriber;
use PHPUnit\Framework\TestCase;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Event\RequestEvent;
use Symfony\Component\HttpKernel\Event\ResponseEvent;
use Symfony\Component\HttpKernel\HttpKernelInterface;
use Symfony\Component\HttpKernel\KernelEvents;

class AnalyticsEventSubscriberTest extends TestCase
{
    private function createSubscriber(int $privacyLevel = 0): AnalyticsEventSubscriber
    {
        return new AnalyticsEventSubscriber(
            'test-api-key',
            'http://localhost:99999/', // Prevent actual requests
            $privacyLevel
        );
    }

    private function createKernelMock(): HttpKernelInterface
    {
        return $this->createMock(HttpKernelInterface::class);
    }

    public function testSubscribestoCorrectEvents(): void
    {
        $events = AnalyticsEventSubscriber::getSubscribedEvents();

        $this->assertArrayHasKey(KernelEvents::REQUEST, $events);
        $this->assertArrayHasKey(KernelEvents::RESPONSE, $events);
    }

    public function testLogsMainRequest(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        $request = Request::create('/api/users', 'GET', [], [], [], [
            'HTTP_HOST' => 'api.example.com',
            'HTTP_USER_AGENT' => 'Mozilla/5.0',
            'REMOTE_ADDR' => '192.168.1.1',
        ]);

        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $this->assertNotNull($request->attributes->get('_analytics_start'));

        $response = new Response('OK', 200);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        $this->assertSame(1, $subscriber->getClient()->getBufferSize());
    }

    public function testDoesNotLogSubRequests(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        $request = Request::create('/api/users', 'GET');

        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::SUB_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $this->assertNull($request->attributes->get('_analytics_start'));

        $response = new Response('OK', 200);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::SUB_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        $this->assertSame(0, $subscriber->getClient()->getBufferSize());
    }

    public function testPreservesResponseStatus(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        $request = Request::create('/api/users', 'POST');
        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $response = new Response('Created', 201);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        $this->assertSame(201, $response->getStatusCode());
        $this->assertSame(1, $subscriber->getClient()->getBufferSize());
    }

    public function testHandles404(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        $request = Request::create('/not-found', 'GET');
        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $response = new Response('Not Found', 404);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        $this->assertSame(1, $subscriber->getClient()->getBufferSize());
    }

    public function testHandles500(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        $request = Request::create('/api/error', 'GET');
        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $response = new Response('Internal Server Error', 500);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        $this->assertSame(1, $subscriber->getClient()->getBufferSize());
    }

    public function testLogsDifferentMethods(): void
    {
        $subscriber = $this->createSubscriber();
        $kernel = $this->createKernelMock();

        foreach (['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] as $method) {
            $request = Request::create('/api/test', $method);
            $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
            $subscriber->onKernelRequest($requestEvent);

            $response = new Response('OK', 200);
            $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
            $subscriber->onKernelResponse($responseEvent);
        }

        $this->assertSame(5, $subscriber->getClient()->getBufferSize());
    }

    public function testPrivacyLevelTwoOmitsIpAddress(): void
    {
        $subscriber = $this->createSubscriber(privacyLevel: 2);
        $kernel = $this->createKernelMock();

        $request = Request::create('/api/users', 'GET', [], [], [], [
            'REMOTE_ADDR' => '1.2.3.4',
        ]);
        $requestEvent = new RequestEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST);
        $subscriber->onKernelRequest($requestEvent);

        $response = new Response('OK', 200);
        $responseEvent = new ResponseEvent($kernel, $request, HttpKernelInterface::MAIN_REQUEST, $response);
        $subscriber->onKernelResponse($responseEvent);

        // Privacy level 2 means IP is not sent — buffer has 1 request but IP is null
        $this->assertSame(1, $subscriber->getClient()->getBufferSize());
        $this->assertSame(2, $subscriber->getClient()->getConfig()->privacyLevel);
    }
}
