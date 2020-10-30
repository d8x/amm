package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Amm",
	Long:  `All software has versions. This is Amm`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ark Mod Manager v0.1")
	},
}
