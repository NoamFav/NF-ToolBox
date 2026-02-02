package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nf-software/nf-toolbox/internal/toolbox/api"
	"github.com/nf-software/nf-toolbox/internal/toolbox/config"
	"github.com/nf-software/nf-toolbox/pkg/license"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tools",
	Long:  `Show all NF Software tools with their license and installation status.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	session, err := config.LoadSession()
	if err != nil || session.Token == "" {
		return fmt.Errorf("not logged in. Run: nf-toolbox login")
	}

	client := api.NewClient()
	tools, err := client.ListTools(session.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch tools: %w", err)
	}

	binDir := license.GetBinDir()

	fmt.Println("Available tools:")
	fmt.Println()

	for _, tool := range tools {
		// Check if installed
		installed := isInstalled(binDir, tool.Name)

		// Status icon
		var status string
		if tool.Owned {
			if installed {
				status = "✓"
			} else {
				status = "○"
			}
		} else {
			status = "🔒"
		}

		// Format price
		price := fmt.Sprintf("$%d/mo", tool.PriceMonthly/100)

		// Print
		fmt.Printf("  %s %s", status, tool.DisplayName)
		if installed {
			fmt.Print(" (installed)")
		}
		if !tool.Owned {
			fmt.Printf(" - %s", price)
		}
		fmt.Println()
		fmt.Printf("    %s\n", tool.Description)
		fmt.Println()
	}

	fmt.Println("Legend:")
	fmt.Println("  ✓ = owned & installed")
	fmt.Println("  ○ = owned, not installed")
	fmt.Println("  🔒 = not licensed")
	fmt.Println()
	fmt.Println("Install a tool: nf-toolbox install <name>")
	fmt.Println("Purchase: https://nf-software.com")

	return nil
}

func isInstalled(binDir, toolName string) bool {
	path := filepath.Join(binDir, toolName)
	_, err := os.Stat(path)
	return err == nil
}
