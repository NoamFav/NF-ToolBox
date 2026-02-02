// Example: How to integrate license checking into your tool
//
// This shows how Zvezda, Iskra, Puls, etc. should check for a valid license
// on startup before running.

package main

import (
	"fmt"
	"os"

	"github.com/nf-software/nf-toolbox/pkg/license"
)

const toolName = "example-tool" // Change to your tool name: "zvezda", "iskra", etc.

func main() {
	// 1. Check license FIRST before doing anything else
	if err := license.CheckEntitlement(toolName); err != nil {
		printLicenseError(err)
		os.Exit(1)
	}

	// 2. Optional: Check for other tools to enable integrations
	integrations := detectIntegrations()

	// 3. Run your actual tool
	run(integrations)
}

func printLicenseError(err error) {
	fmt.Fprintf(os.Stderr, "License error: %v\n\n", err)
	fmt.Fprintln(os.Stderr, "To use this tool:")
	fmt.Fprintln(os.Stderr, "  1. Install NF-ToolBox")
	fmt.Fprintln(os.Stderr, "  2. Run: nf-toolbox login")
	fmt.Fprintln(os.Stderr, "  3. Run: nf-toolbox install "+toolName)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Or purchase at: https://nf-software.com/"+toolName)
}

// Integrations represents optional features based on other owned tools
type Integrations struct {
	HasIskra  bool // Multi-repo features
	HasDerevo bool // AST-aware features
	HasIris   bool // AI features
}

func detectIntegrations() Integrations {
	return Integrations{
		HasIskra:  license.HasAccess("iskra"),
		HasDerevo: license.HasAccess("derevo"),
		HasIris:   license.HasAccess("iris"),
	}
}

func run(i Integrations) {
	fmt.Println("Tool is running!")
	fmt.Printf("Integrations: %+v\n", i)

	if i.HasIskra {
		fmt.Println("- Multi-repo features enabled (Iskra detected)")
	}
	if i.HasDerevo {
		fmt.Println("- AST-aware features enabled (Derevo detected)")
	}
	if i.HasIris {
		fmt.Println("- AI features enabled (Iris detected)")
	}
}
