<?php

declare(strict_types=1);

namespace ApiAnalytics\Laravel\Tests;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use ApiAnalytics\Laravel\AnalyticsMiddleware;
use ApiAnalytics\Laravel\AnalyticsServiceProvider;
use Orchestra\Testbench\TestCase;

class ServiceProviderTest extends TestCase
{
    protected function getPackageProviders($app): array
    {
        return [AnalyticsServiceProvider::class];
    }

    protected function defineEnvironment($app): void
    {
        $app['config']->set('api-analytics.api_key', 'test-api-key');
        $app['config']->set('api-analytics.privacy_level', 1);
        $app['config']->set('api-analytics.server_url', 'https://custom.server.com/');
    }

    public function testConfigIsRegistered(): void
    {
        $config = $this->app->make(Config::class);

        $this->assertInstanceOf(Config::class, $config);
        $this->assertSame(1, $config->privacyLevel);
        $this->assertSame('https://custom.server.com/', $config->serverUrl);
    }

    public function testClientIsRegistered(): void
    {
        $client = $this->app->make(Client::class);

        $this->assertInstanceOf(Client::class, $client);
        $this->assertSame('Laravel', $client->getFramework());
    }

    public function testMiddlewareIsRegistered(): void
    {
        $middleware = $this->app->make(AnalyticsMiddleware::class);

        $this->assertInstanceOf(AnalyticsMiddleware::class, $middleware);
    }

    public function testClientIsSingleton(): void
    {
        $client1 = $this->app->make(Client::class);
        $client2 = $this->app->make(Client::class);

        $this->assertSame($client1, $client2);
    }

    public function testConfigIsSingleton(): void
    {
        $config1 = $this->app->make(Config::class);
        $config2 = $this->app->make(Config::class);

        $this->assertSame($config1, $config2);
    }

    public function testProvidesReturnsCorrectServices(): void
    {
        $provider = new AnalyticsServiceProvider($this->app);
        $provides = $provider->provides();

        $this->assertContains(Client::class, $provides);
        $this->assertContains(Config::class, $provides);
        $this->assertContains(AnalyticsMiddleware::class, $provides);
    }
}
