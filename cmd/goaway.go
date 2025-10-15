package cmd

import (
	"embed"
	app "goaway/backend"
	"goaway/backend/logging"

	"github.com/spf13/cobra"
)

var log = logging.GetLogger()

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func Execute(version, commit, date string, content embed.FS) error {
	versionInfo := VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	return createRootCommand(versionInfo, content).Execute()
}

func createRootCommand(versionInfo VersionInfo, content embed.FS) *cobra.Command {
	flags := NewFlags()

	cmd := &cobra.Command{
		Use:   "goaway",
		Short: "GoAway is a DNS sinkhole with a web interface",
		Run: func(cmd *cobra.Command, args []string) {
			setFlags := flags.GetSetFlags(cmd)

			application := app.New(setFlags, versionInfo.Version, versionInfo.Commit, versionInfo.Date, content)
			if err := application.Start(); err != nil {
				log.Fatal("Application failed to start: %s", err)
			}
		},
	}

	flags.Register(cmd)
	return cmd
}
