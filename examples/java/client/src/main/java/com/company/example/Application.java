package com.company.example;

import com.retailcortex.skills.loader.SkillLoader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Application {

    private static final Logger logger = LoggerFactory.getLogger(Application.class);

    public static String getSkmServerUrl() {
        return System.getProperty("skm.server.url",
                System.getenv().getOrDefault("SKM_SERVER_URL", "http://localhost:8080"));
    }

    public static String getSkmApiKey() {
        return System.getProperty("skm.api.key",
                System.getenv().getOrDefault("SKM_API_KEY", "java-secret-key-99999"));
    }

    public static void main(String[] args) {
        logger.info("Initializing Java Client Standalone Example...");

        String serverUrl = getSkmServerUrl();
        String apiKey = getSkmApiKey();

        logger.info("Loaded System properties: serverUrl={}, apiKey={}", serverUrl, apiKey);

        var parsed = SkillLoader.parseSkillRootUri("github://google/skills@main/tree/main/skills/cloud/gemini-api");
        logger.info("Parsed URI: scheme={}, target={}, ref={}", parsed.scheme(), parsed.target(), parsed.ref());
    }
}
