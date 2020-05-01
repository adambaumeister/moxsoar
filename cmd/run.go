package cmd

import (
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/adambaumeister/moxsoar/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const DEFAULT_PACK = "moxsoar-content"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		pi := pack.GetPackIndex(viper.GetString("contentdir"))
		pi.Get(DEFAULT_PACK, "https://github.com/adambaumeister/moxsoar-content.git")

		// Use the default pack out of the box.
		p, err := pi.GetPackName(DEFAULT_PACK)
		if err != nil {
			log.Fatal("Could not load default pack name %v during startup!", DEFAULT_PACK)
		}
		rc := runner.GetRunConfig(p.FullPath)
		rc.RunAll()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
