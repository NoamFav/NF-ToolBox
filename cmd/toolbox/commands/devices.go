package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nf-software/nf-toolbox/internal/toolbox/api"
	"github.com/nf-software/nf-toolbox/internal/toolbox/config"
	"github.com/nf-software/nf-toolbox/pkg/license"
)

var DevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage activated devices",
	Long:  `List and manage devices where your license is activated.`,
	RunE:  runDevices,
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate <number>",
	Short: "Deactivate a device",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeactivate,
}

func init() {
	DevicesCmd.AddCommand(deactivateCmd)
}

func runDevices(cmd *cobra.Command, args []string) error {
	session, err := config.LoadSession()
	if err != nil || session.Token == "" {
		return fmt.Errorf("not logged in. Run: nf-toolbox login")
	}

	client := api.NewClient()
	resp, err := client.ListDevices(session.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch devices: %w", err)
	}

	// Get current fingerprint
	currentFP, _ := license.ComputeFingerprint()

	fmt.Printf("Your Active Devices (%d/%d):\n\n", resp.Count, resp.MaxDevices)

	for i, device := range resp.Devices {
		num := i + 1
		fmt.Printf("  %d. %s\n", num, device.Nickname)
		fmt.Printf("     Activated: %s\n", device.ActivatedAt.Format("Jan 2, 2006"))

		// Check if this device
		if device.MachineFingerprint == currentFP {
			fmt.Println("     [This device]")
		} else {
			lastSeen := time.Since(device.LastSeen)
			if lastSeen < 24*time.Hour {
				fmt.Printf("     Last seen: %s ago\n", formatDuration(lastSeen))
			} else {
				days := int(lastSeen.Hours() / 24)
				fmt.Printf("     Last seen: %d days ago\n", days)
			}
		}
		fmt.Println()
	}

	if resp.Count > 1 {
		fmt.Println("To deactivate a device:")
		fmt.Println("  nf-toolbox devices deactivate <number>")
	}

	return nil
}

func runDeactivate(cmd *cobra.Command, args []string) error {
	session, err := config.LoadSession()
	if err != nil || session.Token == "" {
		return fmt.Errorf("not logged in. Run: nf-toolbox login")
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid device number")
	}

	client := api.NewClient()
	resp, err := client.ListDevices(session.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch devices: %w", err)
	}

	if num < 1 || num > len(resp.Devices) {
		return fmt.Errorf("invalid device number (1-%d)", len(resp.Devices))
	}

	device := resp.Devices[num-1]

	// Check if this device
	currentFP, _ := license.ComputeFingerprint()
	if device.MachineFingerprint == currentFP {
		return fmt.Errorf("cannot deactivate current device")
	}

	// Confirm
	fmt.Printf("Deactivate \"%s\"?\n", device.Nickname)
	fmt.Print("Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	if strings.TrimSpace(input) != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	// Deactivate
	if err := client.DeactivateDevice(session.Token, device.ID); err != nil {
		return fmt.Errorf("failed to deactivate: %w", err)
	}

	fmt.Printf("\n✓ Device \"%s\" deactivated\n", device.Nickname)
	fmt.Printf("  %d/%d devices now active\n", resp.Count-1, resp.MaxDevices)

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		return fmt.Sprintf("%d minute(s)", mins)
	}
	hours := int(d.Hours())
	return fmt.Sprintf("%d hour(s)", hours)
}
