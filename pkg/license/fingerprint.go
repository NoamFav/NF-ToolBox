package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	appID = "nf-toolbox"
	salt  = "nf-tools-v1-2025" // Change this to your own secret
)

// ComputeFingerprint generates a unique machine identifier
func ComputeFingerprint() (string, error) {
	machineID, err := getMachineID()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %w", err)
	}

	// Hash: appID + salt + machineID
	raw := appID + salt + machineID
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:]), nil
}

func getMachineID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacOSMachineID()
	case "linux":
		if isWSL() {
			// Combine Linux + Windows IDs for WSL
			linuxID, err := getLinuxMachineID()
			if err != nil {
				return "", err
			}
			winID, _ := getWSLWindowsGuid()
			return linuxID + winID, nil
		}
		return getLinuxMachineID()
	case "windows":
		return getWindowsMachineGuid()
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// macOS: Get IOPlatformUUID
func getMacOSMachineID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ioreg failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			// Format: "IOPlatformUUID" = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				return parts[3], nil
			}
		}
	}

	return "", fmt.Errorf("IOPlatformUUID not found")
}

// Linux: Get /etc/machine-id
func getLinuxMachineID() (string, error) {
	// Try standard location first
	data, err := os.ReadFile("/etc/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Fallback to dbus machine-id
	data, err = os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("machine-id not found")
}

// Windows: Get MachineGuid from registry
func getWindowsMachineGuid() (string, error) {
	cmd := exec.Command("reg", "query",
		"HKLM\\SOFTWARE\\Microsoft\\Cryptography",
		"/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("reg query failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "MachineGuid") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[len(parts)-1], nil
			}
		}
	}

	return "", fmt.Errorf("MachineGuid not found")
}

// Check if running in WSL
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// Get Windows GUID from within WSL
func getWSLWindowsGuid() (string, error) {
	// Access Windows registry from WSL
	cmd := exec.Command("/mnt/c/Windows/System32/reg.exe", "query",
		"HKLM\\SOFTWARE\\Microsoft\\Cryptography",
		"/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "MachineGuid") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[len(parts)-1], nil
			}
		}
	}

	return "", fmt.Errorf("MachineGuid not found in WSL")
}

// GetPlatformInfo returns OS and architecture info
func GetPlatformInfo() (platform, arch string) {
	return runtime.GOOS, runtime.GOARCH
}

// GetNickname generates a friendly device name
func GetNickname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s (%s/%s)", hostname, runtime.GOOS, runtime.GOARCH)
}
