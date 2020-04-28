package cmd

import (
	"github.com/adambaumeister/moxsoar/integrations/minemeld"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		m := minemeld.Minemeld{}
		cd := viper.GetString("contentdir")
		m.Start(cd)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
