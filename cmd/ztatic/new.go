package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var newDir string

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Scaffold a new Ztatic project architecture",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("🚀 Scaffolding new Ztatic project: %s\n", projectName)
		
		baseDir := filepath.Join(newDir, projectName)

		// Define the standard architectural directory structure
		dirs := []string{
			"cmd/server",
			"internal/controllers",
			"internal/models",
			"internal/repositories",
			"internal/views/layouts",
			"internal/views/components",
			"assets/css",
			"assets/js",
			"db/migrations",
		}

		// Create directories
		for _, dir := range dirs {
			path := filepath.Join(baseDir, dir)
			if err := os.MkdirAll(path, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", path, err)
				os.Exit(1)
			}
		}

		// Generate main.go stub
		mainContent := `package main

import (
	"log"
	"ztatic-go-framework"
)

func main() {
	app := ztatic.NewSecure()
	
	app.GET("/", func(c *ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic!")
	})

	log.Fatal(app.Start(":8080"))
}
`
		if err := os.WriteFile(filepath.Join(baseDir, "cmd/server", "main.go"), []byte(mainContent), 0644); err != nil {
			fmt.Printf("Error writing main.go: %v\n", err)
		}

		// Generate tools.go to pre-wire and lock generator and framework runtime dependencies
		toolsContent := `//go:build tools

package main

import (
	_ "github.com/a-h/templ"
	_ "ztatic-go-framework/data"
	_ "ztatic-go-framework/realtime"
)
`
		if err := os.WriteFile(filepath.Join(baseDir, "tools.go"), []byte(toolsContent), 0644); err != nil {
			fmt.Printf("Error writing tools.go: %v\n", err)
		}

		// Locate framework root to seed go.sum and configure replace directive
		var frameworkDir string
		var frameworkSumData []byte

		if cwd, err := os.Getwd(); err == nil {
			if modBytes, err := os.ReadFile(filepath.Join(cwd, "go.mod")); err == nil {
				if strings.Contains(string(modBytes), "module ztatic-go-framework") {
					frameworkDir = cwd
				}
			}
		}

		if frameworkDir == "" {
			for _, rel := range []string{".", "..", "../.."} {
				if modBytes, err := os.ReadFile(filepath.Join(rel, "go.mod")); err == nil {
					if strings.Contains(string(modBytes), "module ztatic-go-framework") {
						if abs, err := filepath.Abs(rel); err == nil {
							frameworkDir = abs
							break
						}
					}
				}
			}
		}

		if frameworkDir != "" {
			frameworkSumData, _ = os.ReadFile(filepath.Join(frameworkDir, "go.sum"))
		} else {
			for _, sumPath := range []string{"go.sum", "../go.sum", filepath.Join(baseDir, "../go.sum")} {
				if data, err := os.ReadFile(sumPath); err == nil && len(data) > 0 {
					frameworkSumData = data
					break
				}
			}
		}

		// Pre-seed go.sum before running go mod tidy for offline/air-gapped stability
		if len(frameworkSumData) > 0 {
			if err := os.WriteFile(filepath.Join(baseDir, "go.sum"), frameworkSumData, 0644); err != nil {
				fmt.Printf("Warning: failed to seed go.sum: %v\n", err)
			}
		}

		replacePath := "../"
		if frameworkDir != "" {
			if rel, err := filepath.Rel(baseDir, frameworkDir); err == nil {
				replacePath = rel
			}
		}

		// Generate go.mod
		modContent := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire (\n\tgithub.com/a-h/templ v0.3.1020\n\tztatic-go-framework v0.0.0\n)\n\nreplace ztatic-go-framework => %s\n", projectName, replacePath)
		if err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(modContent), 0644); err != nil {
			fmt.Printf("Error writing go.mod: %v\n", err)
		}

		// Run go mod tidy
		cmdTidy := exec.Command("go", "mod", "tidy")
		cmdTidy.Dir = baseDir
		if err := cmdTidy.Run(); err != nil {
			fmt.Printf("Warning: failed to run 'go mod tidy': %v\n", err)
		}

		fmt.Println("✅ Project scaffolded successfully!")
		fmt.Printf("Next steps:\n  cd %s\n  ztatic dev\n", baseDir)
	},
}

func init() {
	newCmd.Flags().StringVarP(&newDir, "dir", "d", ".", "Base target directory for project scaffolding")
	rootCmd.AddCommand(newCmd)
}
