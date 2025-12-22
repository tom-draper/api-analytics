<?php

declare(strict_types=1);

namespace ApiAnalytics\Laravel;

use ApiAnalytics\Core\Client;
use ApiAnalytics\Core\Config;
use Illuminate\Contracts\Foundation\Application;
use Illuminate\Support\ServiceProvider;

/**
 * Laravel service provider for API Analytics.
 *
 * Registers the Analytics service and middleware with the Laravel container.
 */
class AnalyticsServiceProvider extends ServiceProvider
{
    /**
     * Register the application services.
     */
    public function register(): void
    {
        $this->mergeConfigFrom(
            __DIR__ . '/../config/api-analytics.php',
            'api-analytics'
        );

        $this->app->singleton(Config::class, function (Application $app) {
            $config = new Config();

            $analyticsConfig = $app['config']->get('api-analytics', []);

            if (isset($analyticsConfig['privacy_level'])) {
                $config->privacyLevel = (int) $analyticsConfig['privacy_level'];
            }

            if (isset($analyticsConfig['server_url'])) {
                $config->serverUrl = $analyticsConfig['server_url'];
            }

            // Support custom mapper callbacks from config
            if (isset($analyticsConfig['get_user_id']) && is_callable($analyticsConfig['get_user_id'])) {
                $config->setGetUserId($analyticsConfig['get_user_id']);
            }

            return $config;
        });

        $this->app->singleton(Client::class, function (Application $app) {
            $apiKey = $app['config']->get('api-analytics.api_key', '');
            $config = $app->make(Config::class);

            return new Client($apiKey, 'Laravel', $config);
        });

        $this->app->singleton(AnalyticsMiddleware::class, function (Application $app) {
            return new AnalyticsMiddleware($app->make(Client::class));
        });
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        if ($this->app->runningInConsole()) {
            $this->publishes([
                __DIR__ . '/../config/api-analytics.php' => config_path('api-analytics.php'),
            ], 'api-analytics-config');
        }
    }

    /**
     * Get the services provided by the provider.
     *
     * @return array<string>
     */
    public function provides(): array
    {
        return [
            Client::class,
            Config::class,
            AnalyticsMiddleware::class,
        ];
    }
}
