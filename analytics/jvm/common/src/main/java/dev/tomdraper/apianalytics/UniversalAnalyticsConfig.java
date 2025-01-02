package dev.tomdraper.apianalytics;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import static dev.tomdraper.apianalytics.PayloadHandler.DEFAULT_ENDPOINT;

/** A universal config class that all JVM implementations of ApiAnalytics rely on.
 *  Not that the PlayFramework subproject does not rely on this as a superclass, and instead uses a config file.
 * @see #apiKey
 * @see #timeout
 * @see #privacyLevel
 * @see #serverUrl
 * */
public class UniversalAnalyticsConfig {
    /** The API key for */
    @Nullable
    public String apiKey;
    /** Changes the time between requests to the service in milliseconds. Changing this value is not recommended. */
    @NotNull
    public Long timeout = 60000L;
    //
    @NotNull
    public Integer privacyLevel = 0;
    /** Changes where the plugin sends requests to. Don't change this unless you don't use the main server (i.e. self-hosting or secondary server) */
    @NotNull
    public String serverUrl = DEFAULT_ENDPOINT;
}
