package main

import "github.com/spf13/cobra"

// deepCopy is a helper function for deeply copying a Cobra command.
func deepCopy(cmd *cobra.Command) *cobra.Command {
	newCmd := &cobra.Command{}
	*newCmd = *cmd
	return newCmd
}
