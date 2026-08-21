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

import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.stream.Collectors;

/**
 * High-performance registry for discovering and querying enterprise skills in Java ADK agents.
 */
public class SkillRegistry {

    private final Path root;
    private final Map<String, SkillDefinition> skills;

    public SkillRegistry() {
        this(null, null, null, null);
    }

    public SkillRegistry(Path skillsRoot, List<String> roots, List<String> filter, Path dotenvPath) {
        this.root = skillsRoot != null ? skillsRoot.toAbsolutePath() : SkillLoader.findRegistryRoot();
        List<String> rootList = roots;
        if (rootList == null || rootList.isEmpty()) {
            if (skillsRoot != null) {
                rootList = List.of(skillsRoot.toString());
            }
        }
        this.skills = SkillLoader.loadSkillsFromRoots(rootList, filter, null, dotenvPath);
    }

    public static SkillRegistry fromRoots(List<String> roots, List<String> filter, String token, Path dotenvPath) {
        SkillRegistry registry = new SkillRegistry();
        Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromRoots(roots, filter, token, dotenvPath);
        registry.skills.clear();
        registry.skills.putAll(loaded);
        return registry;
    }

    public static SkillRegistry fromGithub(String repo, String ref, List<String> roots, List<String> filter,
                                            String token, Path dotenvPath) {
        SkillRegistry registry = new SkillRegistry();
        Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromGithub(repo, ref, roots, filter, token, dotenvPath);
        registry.skills.clear();
        registry.skills.putAll(loaded);
        return registry;
    }

    public Map<String, SkillDefinition> getSkills() {
        return Collections.unmodifiableMap(skills);
    }

    public SkillDefinition get(String name) {
        return skills.get(name);
    }

    public List<SkillSummary> listSkills() {
        return skills.values().stream()
                .map(s -> new SkillSummary(
                        s.getName(),
                        s.getDescription(),
                        s.getReferences() != null ? s.getReferences().size() : 0,
                        s.getExamples() != null ? s.getExamples().size() : 0,
                        s.getPath()
                ))
                .sorted(Comparator.comparing(SkillSummary::getName))
                .collect(Collectors.toList());
    }

    public List<SkillDefinition> search(String query) {
        if (query == null || query.isBlank()) {
            return Collections.emptyList();
        }
        String q = query.toLowerCase();
        return skills.values().stream()
                .filter(s -> {
                    boolean match = (s.getName() != null && s.getName().toLowerCase().contains(q))
                            || (s.getDescription() != null && s.getDescription().toLowerCase().contains(q))
                            || (s.getInstructions() != null && s.getInstructions().toLowerCase().contains(q));
                    if (!match && s.getReferences() != null) {
                        match = s.getReferences().values().stream().anyMatch(refText -> refText.toLowerCase().contains(q));
                    }
                    return match;
                })
                .sorted(Comparator.comparing(SkillDefinition::getName))
                .collect(Collectors.toList());
    }

    public List<SkillDefinition> getDomainSkills(String domain) {
        if (domain == null || domain.isBlank()) {
            return Collections.emptyList();
        }
        String domainNorm = domain.toLowerCase();
        return skills.values().stream()
                .filter(s -> (s.getName() != null && s.getName().toLowerCase().contains(domainNorm))
                        || (s.getDescription() != null && s.getDescription().toLowerCase().contains(domainNorm)))
                .sorted(Comparator.comparing(SkillDefinition::getName))
                .collect(Collectors.toList());
    }

    public List<SkillDefinition> suggestSkills(String prompt, int maxSkills, String serverUrl) {
        if (prompt == null || prompt.isBlank()) {
            return skills.values().stream().limit(Math.max(1, maxSkills)).collect(Collectors.toList());
        }
        int boundedMax = Math.min(Math.max(1, maxSkills), 25);
        List<SkillDefinition> matches = search(prompt);
        if (!matches.isEmpty()) {
            return matches.stream().limit(boundedMax).collect(Collectors.toList());
        }
        String[] words = prompt.toLowerCase().split("\\s+");
        Set<String> seen = new HashSet<>();
        List<SkillDefinition> domainMatches = new ArrayList<>();
        for (String w : words) {
            for (SkillDefinition s : getDomainSkills(w)) {
                if (seen.add(s.getName())) {
                    domainMatches.add(s);
                }
            }
        }
        if (!domainMatches.isEmpty()) {
            return domainMatches.stream().limit(boundedMax).collect(Collectors.toList());
        }
        return skills.values().stream().limit(boundedMax).collect(Collectors.toList());
    }

    public Path getRoot() {
        return root;
    }
}
