package cmd

import (
	"github.com/adambaumeister/moxsoar/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const DEFAULT_PACK = "default"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		pi := runner.GetPackIndex(viper.GetString("contentdir"))
		// Use the default pack out of the box.
		p, err := pi.GetPackName(DEFAULT_PACK)
		if err != nil {
			log.Fatal("Could not load default pack name during startup!")
		}
		rc := runner.GetRunConfig(p.FullPath)
		rc.RunAll()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
