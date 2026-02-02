package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/nf-software/nf-toolbox/internal/toolbox/api"
	"github.com/nf-software/nf-toolbox/internal/toolbox/config"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your NF Software account",
	Long:  `Authenticate with your NF Software account to access your licensed tools.`,
	RunE:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check if already logged in
	session, _ := config.LoadSession()
	if session != nil && session.Token != "" {
		fmt.Printf("Already logged in as %s\n", session.Email)
		fmt.Println("Run 'nf-toolbox logout' first to switch accounts.")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	// Get email
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	// Get password (hidden)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	fmt.Println() // newline after hidden input

	// Login
	fmt.Println("Logging in...")
	client := api.NewClient()
	resp, err := client.Login(email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save session
	session = &config.Session{
		Token:  resp.Token,
		UserID: resp.User.ID,
		Email:  resp.User.Email,
	}
	if err := config.SaveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	fmt.Printf("\n✓ Logged in as %s\n", resp.User.Email)

	// Fetch and show owned tools
	tools, err := client.ListTools(resp.Token)
	if err == nil {
		var owned []string
		for _, t := range tools {
			if t.Owned {
				owned = append(owned, t.DisplayName)
			}
		}
		if len(owned) > 0 {
			fmt.Printf("\nYou own: %s\n", strings.Join(owned, ", "))
			fmt.Println("\nInstall your tools:")
			fmt.Println("  nf-toolbox install <tool>")
		} else {
			fmt.Println("\nNo tools licensed yet.")
			fmt.Println("Visit https://nf-software.com to get started.")
		}
	}

	return nil
}

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from your account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.DeleteSession(); err != nil {
			return err
		}
		fmt.Println("✓ Logged out")
		return nil
	},
}
