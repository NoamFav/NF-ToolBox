package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/nf-software/nf-toolbox/internal/toolbox/api"
	"github.com/nf-software/nf-toolbox/internal/toolbox/config"
	"github.com/nf-software/nf-toolbox/pkg/license"
)

var InstallCmd = &cobra.Command{
	Use:   "install <tool>",
	Short: "Install a tool",
	Long:  `Download and install a licensed tool to ~/.nf-tools/bin/`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Check session
	session, err := config.LoadSession()
	if err != nil || session.Token == "" {
		return fmt.Errorf("not logged in. Run: nf-toolbox login")
	}

	client := api.NewClient()

	// Check if user owns this tool
	tools, err := client.ListTools(session.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch tools: %w", err)
	}

	var targetTool *api.Tool
	for _, t := range tools {
		if t.Name == toolName {
			targetTool = &t
			break
		}
	}

	if targetTool == nil {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	if !targetTool.Owned {
		fmt.Printf("You don't own %s\n", targetTool.DisplayName)
		fmt.Printf("Purchase at: https://nf-software.com/%s\n", toolName)
		return nil
	}

	// Check if already installed
	binDir := license.GetBinDir()
	binPath := filepath.Join(binDir, toolName)
	if _, err := os.Stat(binPath); err == nil {
		fmt.Printf("%s is already installed\n", targetTool.DisplayName)
		fmt.Println("Run: nf-toolbox update")
		return nil
	}

	// Compute fingerprint
	fingerprint, err := license.ComputeFingerprint()
	if err != nil {
		return fmt.Errorf("failed to compute fingerprint: %w", err)
	}

	// Activate and get license
	fmt.Printf("Activating %s...\n", targetTool.DisplayName)
	activationResp, err := client.Activate(session.Token, &api.ActivationRequest{
		MachineFingerprint: fingerprint,
		Nickname:           license.GetNickname(),
		Platform:           runtime.GOOS,
		Arch:               runtime.GOARCH,
	})
	if err != nil {
		return fmt.Errorf("activation failed: %w", err)
	}

	// Save license
	if err := license.SaveLicense(activationResp.License); err != nil {
		return fmt.Errorf("failed to save license: %w", err)
	}

	// Download binary
	fmt.Printf("Downloading %s...\n", targetTool.DisplayName)
	downloadURL := getDownloadURL(toolName)

	if err := downloadBinary(downloadURL, binPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Make executable
	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	fmt.Printf("\n✓ %s installed successfully\n", targetTool.DisplayName)
	fmt.Printf("  Location: %s\n", binPath)
	fmt.Println()

	// Check if bin dir is in PATH
	if !isInPath(binDir) {
		fmt.Println("Add to your PATH (one-time):")
		fmt.Printf("  echo 'export PATH=\"%s:$PATH\"' >> ~/.zshrc\n", binDir)
		fmt.Println()
	}

	fmt.Printf("Run: %s\n", toolName)

	return nil
}

func getDownloadURL(toolName string) string {
	// TODO: Get from server or manifest
	baseURL := "https://releases.nf-software.com"
	return fmt.Sprintf("%s/%s/latest/%s-%s-%s",
		baseURL, toolName, toolName, runtime.GOOS, runtime.GOARCH)
}

func downloadBinary(url, destPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// Copy with progress
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)

	_, err = io.Copy(writer, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// TODO: Verify checksum from manifest
	_ = hex.EncodeToString(hash.Sum(nil))

	// Move to final location
	return os.Rename(tmpPath, destPath)
}

func isInPath(dir string) bool {
	path := os.Getenv("PATH")
	paths := filepath.SplitList(path)
	for _, p := range paths {
		if p == dir {
			return true
		}
	}
	return false
}

var UninstallCmd = &cobra.Command{
	Use:   "uninstall <tool>",
	Short: "Uninstall a tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]
		binDir := license.GetBinDir()
		binPath := filepath.Join(binDir, toolName)

		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return fmt.Errorf("%s is not installed", toolName)
		}

		if err := os.Remove(binPath); err != nil {
			return fmt.Errorf("failed to remove: %w", err)
		}

		fmt.Printf("✓ %s uninstalled\n", toolName)
		return nil
	},
}
