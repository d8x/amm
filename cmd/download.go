package cmd

import (
	"fmt"
	"strings"

	"github.com/d8x/amm/pkg/steam"
	"github.com/spf13/cobra"
)

const steamCMDDownload = "steamcmd"

func init() {
	rootCmd.AddCommand(downloadCMD)
}

var downloadCMD = &cobra.Command{
	Use:   "download",
	Short: "download an asset",
	Run: func(cmd *cobra.Command, args []string) {
		steamHandler, err := steam.NewSteamHandler()
		if err != nil {
			fmt.Printf("error while creating dir steam %v\n", err)
			return
		}
		for _, a := range args {
			if strings.EqualFold(a, steamCMDDownload) {
				fmt.Println("not supported")
				return
			} else {
				if err := steamHandler.DownloadMod(a); err != nil {
					fmt.Printf("error while downloading mod: %v\n", err)
				}
			}
		}

		fmt.Printf("Download, args %s\n", args)
	},
}
