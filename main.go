//go:generate go install -v github.com/kevinburke/go-bindata/go-bindata
//go:generate go-bindata -prefix res/ -pkg assets -o assets/assets.go res/Franz.lnk
//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"path"

	"github.com/portapps/franz-portable/assets"
	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/shortcut"
	"github.com/portapps/portapps/v3/pkg/utl"
)

type config struct {
	Cleanup      bool `yaml:"cleanup" mapstructure:"cleanup"`
	DisableDelay bool `yaml:"disable_delay" mapstructure:"disable_delay"`
}

var (
	app *portapps.App
	cfg *config
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup:      false,
		DisableDelay: false,
	}

	// Init app
	if app, err = portapps.NewWithCfg("franz-portable", "Franz", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)
	app.Process = utl.PathJoin(app.AppPath, "Franz.exe")
	app.Args = []string{
		"--user-data-dir=" + app.DataPath,
	}

	// Data dir
	os.Setenv("FRANZ_APPDATA_DIR", app.DataPath)

	// Cleanup on exit
	if cfg.Cleanup {
		defer func() {
			utl.Cleanup([]string{
				path.Join(os.Getenv("APPDATA"), "Franz"),
			})
		}()
	}

	// Copy default shortcut
	shortcutPath := path.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Franz Portable.lnk")
	defaultShortcut, err := assets.Asset("Franz.lnk")
	if err != nil {
		log.Error().Err(err).Msg("Cannot load asset Franz.lnk")
	}
	err = os.WriteFile(shortcutPath, defaultShortcut, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Cannot write default shortcut")
	}

	// Check delay
	if cfg.DisableDelay {
		os.Setenv("FRANZ_DELAY", "false")
	} else {
		os.Setenv("FRANZ_DELAY", "true")
	}

	// Update default shortcut
	err = shortcut.Create(shortcut.Shortcut{
		ShortcutPath:     shortcutPath,
		TargetPath:       app.Process,
		Arguments:        shortcut.Property{Clear: true},
		Description:      shortcut.Property{Value: "Franz Portable by Portapps"},
		IconLocation:     shortcut.Property{Value: app.Process},
		WorkingDirectory: shortcut.Property{Value: app.AppPath},
	})
	if err != nil {
		log.Error().Err(err).Msg("Cannot create shortcut")
	}
	defer func() {
		if err := os.Remove(shortcutPath); err != nil {
			log.Error().Err(err).Msg("Cannot remove shortcut")
		}
	}()

	defer app.Close()
	app.Launch(os.Args[1:])
}
