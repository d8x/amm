package cmd

import (
	"fmt"

	"github.com/d8x/amm/pkg/steam"
	"github.com/spf13/cobra"
)

const steamCMDDownload = "steamcmd"

func init() {
	rootCmd.AddCommand(downloadCMD)
	downloadCMD.Flags().StringSliceP("mods", "m", []string{}, "Set mod ids")
	downloadCMD.Flags().BoolP("unpack", "u", false, "Unpack the mods")
	downloadCMD.Flags().StringP("workdir", "w", "amm-workdir", "Working directory")
}

var downloadCMD = &cobra.Command{
	Use:   "download",
	Short: "download an asset",
	Run: func(cmd *cobra.Command, args []string) {
		workDir, err := cmd.Flags().GetString("workdir")
		if err != nil {
			fmt.Printf("error with workdir %v\n", err)
			return
		}
		steamHandler, err := steam.NewSteamHandler(workDir)
		if err != nil {
			fmt.Printf("error when creating steam handler %v\n", err)
			return
		}
		mods, _ := cmd.Flags().GetStringSlice("mods")
		// unpack, _:= cmd.Flags().GetBool("unpack")
		for _, modID := range mods {
			location, err := steamHandler.DownloadMod(modID)
			if err != nil {
				fmt.Printf("error while downloading mod: %v\n", err)
			}
			fmt.Printf("mod downloaded %s\n", location)
		}
		// for _, a := range args {
		// 	if strings.EqualFold(a, steamCMDDownload) {
		// 		fmt.Println("not supported")
		// 		return
		// 	} else {
		//
		// 	}
		// }

	},
}
