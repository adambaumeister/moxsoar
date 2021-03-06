package cmd

import (
	"fmt"
	"github.com/adambaumeister/moxsoar/api"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/adambaumeister/moxsoar/settings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"path"
)

const DEFAULT_PACK = "moxsoar-content"
const DEFAULT_REPO = "https://github.com/adambaumeister/moxsoar-content.git"

var runCmd = &cobra.Command{
	Use: "run",

	Short: "Start the server.",
	Long:  "Starts all configured Mock engines and content.",
	Run: func(cmd *cobra.Command, args []string) {
		// start the API server first

		pi := pack.GetPackIndex(viper.GetString("contentdir"))
		// Pull the default content repository
		p, err := pi.GetOrClone(DEFAULT_PACK, DEFAULT_REPO)

		if err != nil {
			log.Fatal(fmt.Sprintf("Could not load default pack name %s during startup (%)!", DEFAULT_PACK, err))
		}

		settingsFile := path.Join(viper.GetString("datadir"), "settings.json")
		sdb := settings.SettingsDB{
			Path: settingsFile,
		}
		settings := sdb.GetSettings()
		rc := pack.GetRunConfig(p.Path, settings)
		_, _ = pi.ActivatePack(p.Name)
		rc.RunAll()

		if err := settings.CheckForCerts(); err != nil {
			log.Print("Certificate for management not found, using plaintext (port 8080) instead.")
			api.Start(":8080", pi, rc, path.Join(viper.GetString("datadir")), viper.GetString("staticdir"), "", "")
		} else {
			api.Start(":8443", pi, rc, path.Join(viper.GetString("datadir")), viper.GetString("staticdir"), settings.SSLCertificatePath, settings.SSLKeyPath)
		}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
