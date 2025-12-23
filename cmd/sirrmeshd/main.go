package main

import (
	"fmt"
	"os"

	"github.com/sirrchat/SirrMesh/cmd/sirrmeshd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
