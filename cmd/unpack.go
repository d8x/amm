package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(unpackCMD)
}

var unpackCMD = &cobra.Command{
	Use:   "unpack",
	Short: "unpack an asset",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unpack")
	},
}
