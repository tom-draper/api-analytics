package dev.tomdraper.apianalytics.spring;

import dev.tomdraper.apianalytics.PayloadHandler;
import jakarta.annotation.PreDestroy;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

@Component
public class AnalyticsFilter extends OncePerRequestFilter {
    
    private static final Logger logger = LoggerFactory.getLogger(AnalyticsFilter.class);
    private final SpringAnalyticsConfig config;
    public static SpringAnalyticsHandler handler;
    
    public AnalyticsFilter(SpringAnalyticsConfig config) {
        this.config = config;
        handler = new SpringAnalyticsHandler(
            config.apiKey,
            config.timeout,
            config.serverUrl,
            config.privacyLevel
        );
    }
    
    @Override
    protected void doFilterInternal(
            @NotNull HttpServletRequest request, @NotNull HttpServletResponse response, @NotNull FilterChain filterChain
    ) throws ServletException, IOException {
        long startTime = System.currentTimeMillis();
        
        filterChain.doFilter(request, response);
        
        long endTime = System.currentTimeMillis();
        long elapsedTime = endTime - startTime;
        
        String ipAddress = request.getHeader("X-Forwarded-For");
        if (ipAddress == null || ipAddress.isEmpty()) {
            ipAddress = request.getRemoteAddr();
        }
        
        String userIdString = request.getHeader(config.getUserHeader());
        @Nullable Integer userId = null;
        try {
            if (config.getSendUserId() && userIdString != null) {
                userId = Integer.parseInt(userIdString);
            }
        } catch (NumberFormatException e) {
            logger.error("Invalid User ID! Defaulting to sending null", e);
        }
    
        PayloadHandler.RequestData reqData = new PayloadHandler.RequestData(
                request.getServerName(),
                ipAddress,
                request.getHeader("User-Agent"),
                request.getContextPath(),
                response.getStatus(),
                request.getMethod(),
                elapsedTime,
                userId,
                PayloadHandler.createdAt()
        );
        
        handler.logRequest(reqData);
    }
    
    @PreDestroy
    public void destroy() {
        handler.forceSend();
    }
}

