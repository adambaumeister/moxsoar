package cmd

import (
	"github.com/adambaumeister/moxsoar/integrations/minemeld"
	"github.com/adambaumeister/moxsoar/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		m := minemeld.Minemeld{}
		rc := runner.GetRunConfig(viper.GetString("contentdir"))
		m.Start(rc)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
