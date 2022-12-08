<?php
 
namespace App\Http\Middleware;
 
use Closure;
 
class Analytics {
    public function handle($request, Closure $next) {
        $start = microtime(true);
        $response = $next($request);
 
        $data = array(
            'api_key' => '',
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
    
        $context = stream_context_create(array(
            'http' => array(
                'header'  => "Content-type: application/json\r\n",
                'method'  => 'POST',
                'content' => json_encode($data),
                // 'timeout' => .01  // Don't wait for response
            )
            ));
        try {
            file_get_contents($url, false, $context);
        } catch( Exception $e ){
            // Fail silently
        }
    }
}