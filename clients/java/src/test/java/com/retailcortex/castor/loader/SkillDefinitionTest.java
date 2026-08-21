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

package com.retailcortex.castor.loader;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

@DisplayName("SkillDefinition Unit Tests")
class SkillDefinitionTest {

    @Test
    @DisplayName("Should correctly serialize toMap matching Python/Go structure")
    void testToMapSerialization() {
        SkillDefinition skill = new SkillDefinition(
                "java-test-skill",
                "Java client test skill",
                "Follow Java 17 guidelines.",
                "Apache-2.0",
                "Retail Cortex",
                "1.0.0",
                "ADK-v2",
                Map.of("custom_key", "custom_val"),
                Map.of("ref1.md", "Ref content"),
                Map.of("Main.java", "Main content"),
                "/path/to/java/skill"
        );

        Map<String, Object> map = skill.toMap();
        assertThat(map.get("name")).isEqualTo("java-test-skill");
        assertThat(map.get("description")).isEqualTo("Java client test skill");
        assertThat(map.get("instructions")).isEqualTo("Follow Java 17 guidelines.");
        assertThat(map.get("license")).isEqualTo("Apache-2.0");
        assertThat(map.get("author")).isEqualTo("Retail Cortex");
        assertThat(map.get("version")).isEqualTo("1.0.0");
        assertThat(map.get("compatibility")).isEqualTo("ADK-v2");

        @SuppressWarnings("unchecked")
        Map<String, String> meta = (Map<String, String>) map.get("metadata");
        assertThat(meta).containsEntry("custom_key", "custom_val");

        @SuppressWarnings("unchecked")
        List<String> refs = (List<String>) map.get("references");
        assertThat(refs).containsExactly("ref1.md");

        @SuppressWarnings("unchecked")
        List<String> examples = (List<String>) map.get("examples");
        assertThat(examples).containsExactly("Main.java");
    }
}
