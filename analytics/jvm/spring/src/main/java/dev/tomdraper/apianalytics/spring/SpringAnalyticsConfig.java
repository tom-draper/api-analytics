package dev.tomdraper.apianalytics.spring;

import dev.tomdraper.apianalytics.UniversalAnalyticsConfig;
import lombok.Getter;
import lombok.Setter;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.springframework.boot.context.properties.ConfigurationProperties;

@SuppressWarnings("unused")
@ConfigurationProperties(prefix = "dev.tomd.analytics")
@Getter @Setter
public class SpringAnalyticsConfig extends UniversalAnalyticsConfig {
    private @NotNull Boolean sendUserId = false;
    private @Nullable String userHeader;
}
