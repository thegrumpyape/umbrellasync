package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thegrumpyape/umbrellasync/cmd/config"
	"github.com/thegrumpyape/umbrellasync/cmd/sync"
	"github.com/thegrumpyape/umbrellasync/cmd/version"
	"github.com/thegrumpyape/umbrellasync/pkg/configurationManager"
	"github.com/thegrumpyape/umbrellasync/pkg/logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

var debug bool
var logger logging.Logger
var rootCmd = &cobra.Command{
	Use:   "umbrellasync",
	Short: "Syncing threat intel with Umbrella",
	Long:  "Syncing threat intel with Umbrella",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Error(err)
	}
}

var CliVersion = "0.0.1"

func init() {
	lumberjackLogrotate := &lumberjack.Logger{
		Filename:   "./logs/umbrellasync.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.TextFormatter{ForceColors: true},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
	fileHook := logging.NewFileHook(lumberjackLogrotate, &logrus.JSONFormatter{})
	logger.AddHook(fileHook)
	configurationManager := configurationManager.New()
	cobra.OnInitialize(configurationManager.InitConfigFile)
	rootCmd.AddCommand(sync.New(&sync.SyncCommandDependencies{
		ConfigurationManager: configurationManager,
		Logger:               logger,
	}))
	rootCmd.AddCommand(config.New(&config.ConfigCommandDependencies{
		ConfigurationManager: configurationManager,
	}))

	rootCmd.AddCommand(version.New(&version.VersionCommandDependencies{
		CliVersion: CliVersion,
	}))

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if debug {
			logger.SetLevel(logrus.DebugLevel)
		}
	}

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
}
