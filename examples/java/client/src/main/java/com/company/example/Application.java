// Copyright 2026 Ryan McGuinness
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.company.example;

import com.retailcortex.skills.loader.SkillLoader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Application {

    private static final Logger logger = LoggerFactory.getLogger(Application.class);

    public static String getCastorServerUrl() {
        String prop = System.getProperty("castor.server.url", System.getProperty("skm.server.url"));
        if (prop != null && !prop.isBlank()) {
            return prop;
        }
        String env = System.getenv("CASTOR_SERVER_URL");
        if (env == null || env.isBlank()) {
            env = System.getenv("CSTR_SERVER_URL");
        }
        if (env == null || env.isBlank()) {
            env = System.getenv("SKM_SERVER_URL");
        }
        return env != null && !env.isBlank() ? env : "http://localhost:8080";
    }

    public static String getCastorApiKey() {
        String prop = System.getProperty("castor.api.key", System.getProperty("skm.api.key"));
        if (prop != null && !prop.isBlank()) {
            return prop;
        }
        String env = System.getenv("CASTOR_API_KEY");
        if (env == null || env.isBlank()) {
            env = System.getenv("CSTR_API_KEY");
        }
        if (env == null || env.isBlank()) {
            env = System.getenv("SKM_API_KEY");
        }
        return env != null && !env.isBlank() ? env : "java-secret-key-99999";
    }

    public static String getSkmServerUrl() {
        return getCastorServerUrl();
    }

    public static String getSkmApiKey() {
        return getCastorApiKey();
    }

    public static void main(String[] args) {
        logger.info("Initializing Castor Java Client Standalone Example...");

        String serverUrl = getCastorServerUrl();
        String apiKey = getCastorApiKey();

        logger.info("Loaded System properties: serverUrl={}, apiKey={}", serverUrl, apiKey);

        var parsed = SkillLoader.parseSkillRootUri("castor://skills/example.com/testing/test-skill/1.0.0");
        logger.info("Parsed URI: scheme={}, target={}, ref={}", parsed.scheme(), parsed.target(), parsed.ref());
    }
}
