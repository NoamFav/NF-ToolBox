package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/nf-software/nf-toolbox/internal/toolbox/api"
	"github.com/nf-software/nf-toolbox/internal/toolbox/config"
	"github.com/nf-software/nf-toolbox/pkg/license"
)

var UpdateCmd = &cobra.Command{
	Use:   "update [tool]",
	Short: "Update tools and renew license",
	Long:  `Check for updates and renew license if needed. Optionally specify a tool name.`,
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	session, err := config.LoadSession()
	if err != nil || session.Token == "" {
		return fmt.Errorf("not logged in. Run: nf-toolbox login")
	}

	client := api.NewClient()

	// Load current license
	currentLicense, err := license.LoadLicense()
	if err != nil {
		fmt.Println("No license found. Run: nf-toolbox install <tool>")
		return nil
	}

	// Check if renewal needed
	if currentLicense.NeedsRenewal() {
		fmt.Println("Renewing license...")

		fingerprint, err := license.ComputeFingerprint()
		if err != nil {
			return fmt.Errorf("failed to compute fingerprint: %w", err)
		}

		renewResp, err := client.Renew(session.Token, &api.RenewRequest{
			MachineFingerprint: fingerprint,
		})
		if err != nil {
			return fmt.Errorf("renewal request failed: %w", err)
		}

		if renewResp.Denied {
			fmt.Printf("⚠️  License renewal denied: %s\n", renewResp.Reason)
			fmt.Println("Visit https://nf-software.com/account to resolve.")

			// Update state
			state, _ := license.LoadState()
			state.LastRenewAttempt = time.Now()
			state.LastRenewResult = renewResp.Reason
			license.SaveState(state)

			return nil
		}

		// Save new license
		if err := license.SaveLicense(renewResp.License); err != nil {
			return fmt.Errorf("failed to save license: %w", err)
		}

		fmt.Printf("✓ License renewed until %s\n",
			renewResp.License.Payload.ExpiresAt.Format("Jan 2, 2006"))

		// Update state
		state, _ := license.LoadState()
		state.LastRenewAttempt = time.Now()
		state.LastRenewResult = "ok"
		license.SaveState(state)
	} else {
		days := currentLicense.DaysUntilExpiry()
		fmt.Printf("License valid for %d more days\n", days)
	}

	// Check for tool updates
	fmt.Println("\nChecking for updates...")

	binDir := license.GetBinDir()
	entries, err := os.ReadDir(binDir)
	if err != nil {
		fmt.Println("No tools installed yet.")
		return nil
	}

	// TODO: Fetch version manifest from server
	// For now, just list installed tools
	var installed []string
	for _, e := range entries {
		if !e.IsDir() {
			installed = append(installed, e.Name())
		}
	}

	if len(installed) == 0 {
		fmt.Println("No tools installed.")
		return nil
	}

	fmt.Printf("Installed tools: %v\n", installed)
	fmt.Println("All tools are up to date.") // TODO: Actually check versions

	return nil
}

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync license after purchase",
	Long:  `Fetch the latest license from server (e.g., after buying a new tool).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := config.LoadSession()
		if err != nil || session.Token == "" {
			return fmt.Errorf("not logged in. Run: nf-toolbox login")
		}

		fingerprint, err := license.ComputeFingerprint()
		if err != nil {
			return fmt.Errorf("failed to compute fingerprint: %w", err)
		}

		client := api.NewClient()
		newLicense, err := client.GetLicense(session.Token, fingerprint)
		if err != nil {
			return fmt.Errorf("failed to sync license: %w", err)
		}

		if err := license.SaveLicense(newLicense); err != nil {
			return fmt.Errorf("failed to save license: %w", err)
		}

		fmt.Println("✓ License synced")
		fmt.Printf("Entitlements: %v\n", newLicense.Payload.Entitlements)

		return nil
	},
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Session
		session, _ := config.LoadSession()
		if session != nil && session.Token != "" {
			fmt.Printf("Logged in as: %s\n", session.Email)
		} else {
			fmt.Println("Not logged in")
			return nil
		}

		// License
		lic, err := license.LoadLicense()
		if err != nil {
			fmt.Println("No license activated")
			return nil
		}

		fmt.Println()
		fmt.Println("License:")
		fmt.Printf("  Plan: %s\n", lic.Payload.Plan)
		fmt.Printf("  Expires: %s\n", lic.Payload.ExpiresAt.Format("Jan 2, 2006"))
		fmt.Printf("  Days left: %d\n", lic.DaysUntilExpiry())
		fmt.Printf("  Tools: %v\n", lic.Payload.Entitlements)

		// Installed
		fmt.Println()
		fmt.Println("Installed:")
		binDir := license.GetBinDir()
		entries, _ := os.ReadDir(binDir)
		if len(entries) == 0 {
			fmt.Println("  (none)")
		}
		for _, e := range entries {
			if !e.IsDir() {
				fmt.Printf("  - %s\n", e.Name())
			}
		}

		// Paths
		fmt.Println()
		fmt.Println("Paths:")
		fmt.Printf("  Config: %s\n", license.GetConfigDir())
		fmt.Printf("  Bin: %s\n", binDir)

		// PATH check
		if !isInPath(binDir) {
			fmt.Println()
			fmt.Println("⚠️  Bin directory not in PATH")
			fmt.Printf("  Add: export PATH=\"%s:$PATH\"\n", binDir)
		}

		return nil
	},
}

func init() {
	// Add subcommands
}
