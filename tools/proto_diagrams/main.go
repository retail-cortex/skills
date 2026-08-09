package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	workspaceRoot := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if workspaceRoot == "" {
		workspaceRoot = cwd
	}

	buildDir := filepath.Join(workspaceRoot, ".tmp_proto_diagrams")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		log.Fatalf("failed to create temp build dir: %v", err)
	}

	genBinary := filepath.Join(buildDir, "protoc-gen-md-diagrams")
	if _, err := os.Stat(genBinary); os.IsNotExist(err) {
		fmt.Println("Cloning and building GoogleCloudPlatform/proto-gen-md-diagrams...")
		srcDir := filepath.Join(buildDir, "src")
		_ = os.RemoveAll(srcDir)

		cmdClone := exec.Command("git", "clone", "--depth", "1", "https://github.com/GoogleCloudPlatform/proto-gen-md-diagrams.git", srcDir)
		cmdClone.Stdout = os.Stdout
		cmdClone.Stderr = os.Stderr
		if err := cmdClone.Run(); err != nil {
			log.Fatalf("failed to clone proto-gen-md-diagrams: %v", err)
		}

		cmdBuild := exec.Command("go", "build", "-o", genBinary, ".")
		cmdBuild.Dir = srcDir
		cmdBuild.Stdout = os.Stdout
		cmdBuild.Stderr = os.Stderr
		if err := cmdBuild.Run(); err != nil {
			log.Fatalf("failed to build proto-gen-md-diagrams: %v", err)
		}

		_ = os.RemoveAll(srcDir)
	}

	protoDirs := []string{
		filepath.Join(workspaceRoot, "proto", "retailcortex", "skills", "v1"),
		filepath.Join(workspaceRoot, "proto", "retailcortex", "registration", "v1"),
	}
	outDir := filepath.Join(workspaceRoot, "docs", "architecture")

	for _, protoDir := range protoDirs {
		fmt.Printf("Generating Mermaid architecture diagrams from %s/*.proto...\n", protoDir)
		cmdGen := exec.Command(genBinary, "-d", protoDir, "-o", outDir, "-md", "-r=false")
		cmdGen.Stdout = os.Stdout
		cmdGen.Stderr = os.Stderr
		if err := cmdGen.Run(); err != nil {
			log.Fatalf("diagram generation failed for %s: %v", protoDir, err)
		}
	}

	fmt.Println("Protobuf diagrams generated successfully in docs/architecture/")
}
