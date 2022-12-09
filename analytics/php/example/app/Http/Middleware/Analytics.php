<?php
 
namespace App\Http\Middleware;
 
use Closure;
 
class Analytics {

    public function __construct() {
        $this->api_key = getenv("ANALYTICS_API_KEY");
    }
    public function handle($request, Closure $next) {
        $start = microtime(true);
        $response = $next($request);

        $data = array(
            'api_key' => $this->api_key,
            'host' => $request->host(),
            'path' => $request->path(),
            'method' => $request->method(),
            'user_agent' => $request->userAgent(),
            'status' => $response->status(),
            'framework' => 'Laravel',
            'response_time' => (int) ($start - microtime(true)) * 1000,
        );

        $this->log_request($data);
        
        return $response;
    }
    
    private function log_request($data) {
        $url = "http://localhost:8080/api/log-request";
    
        $options = array(
                'http' => array(
                    'header'  => "Content-type: application/json\r\n",
                    'method'  => 'POST',
                    'content' => json_encode($data),
                    // 'timeout' => .01  // Don't wait for response
                )
        );
        
        try {
            $context = stream_context_create($options);
            if (($result = file_get_contents($url, false, $context)) === false) {
                var_dump($http_response_header);
            }
        }  catch (Exception $e) {
            // Fails silently
        }
    }
}