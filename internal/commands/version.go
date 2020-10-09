package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "Print version information about this package",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("OAuth2 Authenticating Proxy v0.1 -- HEAD")
	},
}
