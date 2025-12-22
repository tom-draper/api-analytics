<?php

declare(strict_types=1);

namespace ApiAnalytics\Laravel;

use ApiAnalytics\Core\Client;
use Illuminate\Support\Facades\Facade as BaseFacade;

/**
 * Facade for accessing the Analytics Client.
 *
 * @method static void logRequest(\ApiAnalytics\Core\RequestData $requestData)
 * @method static \ApiAnalytics\Core\RequestData createRequestData(array $context, int $responseTime, int $status)
 * @method static void flush()
 * @method static int getBufferSize()
 * @method static \ApiAnalytics\Core\Config getConfig()
 * @method static string getFramework()
 *
 * @see \ApiAnalytics\Core\Client
 */
class Facade extends BaseFacade
{
    /**
     * Get the registered name of the component.
     */
    protected static function getFacadeAccessor(): string
    {
        return Client::class;
    }
}
