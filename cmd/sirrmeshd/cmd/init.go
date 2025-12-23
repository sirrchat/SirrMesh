/*
SirrMesh - Composable all-in-one email server.
Copyright © 2019-2020 Max Mazurov <fox.cpp@disroot.org>, SirrMesh contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command for initializing sirrmeshd configuration
func NewInitCmd() *cobra.Command {
	var force bool

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize sirrmeshd configuration directory and default config file",
		Long: `Initialize sirrmeshd by creating the configuration directory and
generating a default configuration file.

The configuration directory defaults to ~/.sirrmeshd or can be set via
the MAILCHAT_HOME environment variable.

Example:
  sirrmeshd init
  MAILCHAT_HOME=/etc/sirrmeshd sirrmeshd init
  sirrmeshd init --force  # Overwrite existing config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force)
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration file")

	return initCmd
}

func runInit(force bool) error {
	// Get config directory
	configDir := ConfigDirectory
	configFile := filepath.Join(configDir, "sirrmeshd.conf")

	fmt.Printf("Initializing sirrmeshd...\n")
	fmt.Printf("Configuration directory: %s\n", configDir)

	// Create config directory if not exists
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("✓ Configuration directory created\n")

	// Check if config file already exists
	if _, err := os.Stat(configFile); err == nil {
		if !force {
			fmt.Printf("Configuration file already exists: %s\n", configFile)
			fmt.Printf("Use --force to overwrite\n")
			return nil
		}
		fmt.Printf("Overwriting existing configuration file...\n")
	}

	// Write default config file
	defaultConfig := generateMailConfigContent()
	if err := os.WriteFile(configFile, []byte(defaultConfig), 0o644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	fmt.Printf("✓ Configuration file created: %s\n", configFile)

	// Create additional directories
	dirs := []string{
		filepath.Join(configDir, "dkim_keys"),
		filepath.Join(configDir, "mtasts_cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("✓ Created directory: %s\n", dir)
	}

	fmt.Printf("\nInitialization complete!\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Edit the configuration file: %s\n", configFile)
	fmt.Printf("2. Update hostname and domain settings\n")
	fmt.Printf("3. Configure TLS/ACME settings with your DNS provider\n")
	fmt.Printf("4. Run 'sirrmeshd run' to start the server\n")

	return nil
}
