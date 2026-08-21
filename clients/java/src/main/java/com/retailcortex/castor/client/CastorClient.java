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

package com.retailcortex.castor.client;

import com.retailcortex.castor.loader.SkillDefinition;
import com.retailcortex.castor.loader.SkillLoader;
import com.retailcortex.castor.loader.SkillRegistry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.file.Path;
import java.util.List;
import java.util.Map;

/**
 * CastorClient provides high-level client access to the Castor Registry and local skill repositories.
 */
public class CastorClient {

    private static final Logger logger = LoggerFactory.getLogger(CastorClient.class);

    private final String serverUrl;
    private final String apiKey;
    private final SkillRegistry registry;

    public CastorClient() {
        this(resolveDefaultServerUrl(), resolveDefaultApiKey());
    }

    public CastorClient(String serverUrl, String apiKey) {
        this.serverUrl = serverUrl != null ? serverUrl : "http://localhost:8080";
        this.apiKey = apiKey != null ? apiKey : "";
        this.registry = new SkillRegistry();
    }

    public static String resolveDefaultServerUrl() {
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
        return (env != null && !env.isBlank()) ? env : "http://localhost:8080";
    }

    public static String resolveDefaultApiKey() {
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
        return (env != null && !env.isBlank()) ? env : "";
    }

    public String getServerUrl() {
        return serverUrl;
    }

    public String getApiKey() {
        return apiKey;
    }

    public SkillRegistry getRegistry() {
        return registry;
    }

    public List<SkillDefinition> suggestSkills(String prompt, int maxSkills) {
        return registry.suggestSkills(prompt, maxSkills, serverUrl);
    }

    public SkillLoader.ParsedUri parseUri(String uri) {
        return SkillLoader.parseSkillRootUri(uri);
    }

    public Map<String, SkillDefinition> loadSkillsFromPackage(String packageName, List<String> skillFilter) {
        return SkillLoader.loadSkillsFromPackage(packageName, skillFilter);
    }

    public Map<String, SkillDefinition> loadSkillsFromManifest(Path manifestPath) {
        return SkillLoader.loadSkillsFromManifest(manifestPath);
    }
}
