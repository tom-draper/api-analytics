package dev.tomdraper.apianalytics.spring;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@SuppressWarnings("unused")
@Configuration
@EnableConfigurationProperties(SpringAnalyticsConfig.class)
public class AnalyticsFilterConfiguration { /* no-op */ }
