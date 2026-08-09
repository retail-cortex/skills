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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
)

func main() {
	listFlag := flag.Bool("list", false, "List all registered enterprise skills")
	searchFlag := flag.String("search", "", "Search skills by keyword query")
	getFlag := flag.String("get", "", "Get detailed skill definition by name")
	manifestFlag := flag.Bool("build-manifest", false, "Build pre-compiled JSON skills manifest")
	outputFlag := flag.String("out", "", "Output path for generated manifest JSON")

	flag.Parse()

	registry, err := skillsloader.NewSkillRegistry("", nil, nil, "")
	if err != nil {
		log.Fatalf("Failed to initialize skill registry: %v", err)
	}

	if *manifestFlag {
		outPath, err := skillsloader.BuildSkillsManifest("", *outputFlag)
		if err != nil {
			log.Fatalf("Failed to build skills manifest: %v", err)
		}
		fmt.Printf("Successfully generated skills manifest at: %s\n", outPath)
		return
	}

	if *getFlag != "" {
		skill := registry.Get(*getFlag)
		if skill == nil {
			log.Fatalf("Skill '%s' not found in registry", *getFlag)
		}
		fmt.Printf("Skill: %s\nDescription: %s\nPath: %s\nReferences: %d\nExamples: %d\n\nInstructions:\n%s\n",
			skill.Name, skill.Description, skill.Path, len(skill.References), len(skill.Examples), skill.Instructions)
		return
	}

	if *searchFlag != "" {
		results := registry.Search(*searchFlag)
		fmt.Printf("Search query: '%s' (%d matches found)\n\n", *searchFlag, len(results))
		for _, s := range results {
			fmt.Printf("- %-25s %s\n", s.Name, s.Description)
		}
		return
	}

	if *listFlag || len(os.Args) == 1 {
		summaries := registry.ListSkills()
		fmt.Printf("Enterprise Skills Registry (%d skills loaded)\n", len(summaries))
		fmt.Println(strings.Repeat("=", 60))
		for _, s := range summaries {
			fmt.Printf("- %-25s refs:%d ex:%d path:%s\n", s.Name, s.ReferenceCount, s.ExampleCount, s.Path)
		}
	}
}
