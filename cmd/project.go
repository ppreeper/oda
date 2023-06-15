package cmd

import (
	"github.com/spf13/cobra"
)

// Project Commands (Possible Destructive)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "project commands for init/destruction",
}
