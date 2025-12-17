<?php

return [
    /*
    |--------------------------------------------------------------------------
    | API Analytics API Key
    |--------------------------------------------------------------------------
    |
    | Your unique API key for API Analytics. Generate one at:
    | https://apianalytics.dev/generate
    |
    */
    'api_key' => env('API_ANALYTICS_KEY', ''),

    /*
    |--------------------------------------------------------------------------
    | Privacy Level
    |--------------------------------------------------------------------------
    |
    | Controls client identification by IP address.
    |
    | 0 - IP stored, location inferred and stored (default)
    | 1 - IP used for location inference only, then discarded
    | 2 - IP never sent, location not inferred
    |
    */
    'privacy_level' => env('API_ANALYTICS_PRIVACY_LEVEL', 0),

    /*
    |--------------------------------------------------------------------------
    | Server URL
    |--------------------------------------------------------------------------
    |
    | The API Analytics server URL. Only change this if you're self-hosting.
    |
    */
    'server_url' => env('API_ANALYTICS_SERVER_URL', 'https://www.apianalytics-server.com/'),

    /*
    |--------------------------------------------------------------------------
    | Custom User ID Mapper
    |--------------------------------------------------------------------------
    |
    | Optional callback to extract a custom user ID from the request context.
    | The callback receives an array with 'request' key containing the
    | Illuminate\Http\Request instance.
    |
    | Example:
    |   'get_user_id' => function (array $context) {
    |       $request = $context['request'] ?? null;
    |       return $request?->user()?->id ?? $request?->header('X-API-Key');
    |   },
    |
    */
    'get_user_id' => null,
];
