package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	
	app.GET("/", func(c ztatic.Context) error {
		return c.String(200, "Welcome to Ztatic!")
	})

	log.Fatal(app.Start(":8080"))
}
`
		if err := os.WriteFile(filepath.Join(baseDir, "cmd/server", "main.go"), []byte(mainContent), 0644); err != nil {
			fmt.Printf("Error writing main.go: %v\n", err)
		}

		// Generate go.mod
		modContent := fmt.Sprintf("module %s\n\ngo 1.21\n", projectName)
		if err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(modContent), 0644); err != nil {
			fmt.Printf("Error writing go.mod: %v\n", err)
		}

		fmt.Println("✅ Project scaffolded successfully!")
		fmt.Printf("Next steps:\n  cd %s\n  ztatic dev\n", baseDir)
	},
}

func init() {
	newCmd.Flags().StringVarP(&newDir, "dir", "d", ".", "Base target directory for project scaffolding")
	rootCmd.AddCommand(newCmd)
}
