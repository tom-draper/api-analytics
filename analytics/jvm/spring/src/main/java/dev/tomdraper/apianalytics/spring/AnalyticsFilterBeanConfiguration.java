package dev.tomdraper.apianalytics.spring;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@SuppressWarnings("unused")
@Configuration
@EnableConfigurationProperties({SpringAnalyticsConfig.class})
public class AnalyticsFilterBeanConfiguration {
    @Bean
    public FilterRegistrationBean<AnalyticsFilter> analyticsFilterRegistration(SpringAnalyticsConfig config) {
        FilterRegistrationBean<AnalyticsFilter> registrationBean = new FilterRegistrationBean<>();
        registrationBean.setFilter(new AnalyticsFilter(config));
        registrationBean.addUrlPatterns("/*");
        return registrationBean;
    }
}
