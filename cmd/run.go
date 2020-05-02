package cmd

import (
	"github.com/adambaumeister/moxsoar/api"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/adambaumeister/moxsoar/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const DEFAULT_PACK = "moxsoar-content"
const DEFAULT_REPO = "https://github.com/adambaumeister/moxsoar-content.git"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		// start the API server first

		pi := pack.GetPackIndex(viper.GetString("contentdir"))
		// Pull the default content repository
		p, err := pi.GetOrClone(DEFAULT_PACK, DEFAULT_REPO)

		if err != nil {
			log.Fatal("Could not load default pack name %v during startup!", DEFAULT_PACK)
		}
		rc := runner.GetRunConfig(p.Path)
		rc.RunAll()

		api.Start(":8080", pi, viper.GetString("datadir"))

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
