package dev.tomdraper.apianalytics.spring;

import lombok.Getter;
import lombok.Setter;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.springframework.boot.context.properties.ConfigurationProperties;

import static dev.tomdraper.apianalytics.PayloadHandler.DEFAULT_ENDPOINT;

@SuppressWarnings("unused")
@ConfigurationProperties(prefix = "apianalytics")
@Getter @Setter
public class SpringAnalyticsConfig {
    //
    private @NotNull Boolean sendUserId = false;
    //
    private @Nullable String userHeader;
    //
    public @Nullable String apiKey;
    /** Changes the time between requests to the service in milliseconds. Changing this value is not recommended. */
    public @NotNull Long timeout = 60000L;
    //
    public @NotNull Integer privacyLevel = 0;
    /** Changes where the plugin sends requests to. Don't change this unless you don't use the main server (i.e. self-hosting or secondary server) */
    public @NotNull String serverUrl = DEFAULT_ENDPOINT;
}
