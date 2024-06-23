/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package main

import (
	"fmt"
	"strings"

	stdLog "log"
	"os"
	"time"

	"golang.org/x/exp/slices"

	zeroLog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/kubefirst/kubefirst/cmd"
	"github.com/kubefirst/kubefirst/internal/common"
	"github.com/kubefirst/kubefirst/internal/progress"
	"github.com/kubefirst/runtime/configs"
	"github.com/kubefirst/runtime/pkg"
	"github.com/spf13/viper"
)

func main() {
	bubbleTeaBlacklist := []string{"completion", "help", "--help", "-h", "quota", "logs"}
	canRunBubbleTea := true

	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Info().Msg(err.Error())
	}

	k1Dir := fmt.Sprintf("%s/.k1", homePath)

	//* create k1Dir if it doesn't exist
	if _, err := os.Stat(k1Dir); os.IsNotExist(err) {
		err := os.MkdirAll(k1Dir, os.ModePerm)
		if err != nil {
			log.Info().Msgf("%s directory already exists, continuing", k1Dir)
		}
	}

	//* create config directory
	configFolder := fmt.Sprintf("%s/configs", k1Dir)
	_ = os.Mkdir(configFolder, 0700)
	if err != nil {
		log.Fatal().Msgf("error creating config directory: %s", err)
	}

	//* create log directory
	logsFolder := fmt.Sprintf("%s/logs", k1Dir)
	_ = os.Mkdir(logsFolder, 0700)
	if err != nil {
		log.Fatal().Msgf("error creating logs directory: %s", err)
	}

	config := configs.ReadConfig()

	if config.ConfigName != "" {
		common.ConfigName = config.ConfigName
	}

	argsWithProg := os.Args

	for i := 1; i < len(argsWithProg); i++ {
		arg := os.Args[i]
		isBlackListed := slices.Contains(bubbleTeaBlacklist, arg)

		if isBlackListed {
			canRunBubbleTea = false
		}

		// Check if the argument starts with "--config-name"
		if arg == "--config-name" {
			if i+1 < len(os.Args) {
				common.ConfigName = os.Args[i+1]
			} else {
				log.Fatal().Msg("No config name found")
			}
		} else if strings.HasPrefix(arg, "--config-name") {
			// Get the value of the config name
			parts := strings.Split(arg, "=")
			common.ConfigName = parts[1]
		}
	}

	config.ConfigName = common.ConfigName
	log.Info().Msg(fmt.Sprintf("Using config: %s", common.ConfigName))

	if err := pkg.SetupViper(config, true); err != nil {
		stdLog.Panic(err)
	}

	now := time.Now()
	epoch := now.Unix()
	logfileName := fmt.Sprintf("log_%d.log", epoch)

	isProvision := slices.Contains(argsWithProg, "create")
	isLogs := slices.Contains(argsWithProg, "logs")

	// don't create a new log file for logs, using the previous one
	if isLogs {
		logfileName = viper.GetString("k1-paths.log-file-name")
	}

	// use cluster name as filename
	if isProvision {
		clusterName := common.ConfigName
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]

			// Check if the argument is "--cluster-name"
			if arg == "--cluster-name" && i+1 < len(os.Args) {
				// Get the value of the cluster name
				clusterName = os.Args[i+1]
				break
			} else {
				log.Fatal().Msg("No cluster name found")
			}
		}

		logfileName = fmt.Sprintf("log_%s.log", clusterName)
	}

	//* create session log file
	logfile := fmt.Sprintf("%s/%s", logsFolder, logfileName)
	logFileObj, err := pkg.OpenLogFile(logfile)
	if err != nil {
		stdLog.Panicf("unable to store log location, error is: %s - please verify the current user has write access to this directory", err)
	}

	// handle file close request
	defer func(logFileObj *os.File) {
		err = logFileObj.Close()
		if err != nil {
			log.Print(err)
		}
	}(logFileObj)

	// setup default logging
	// this Go standard log is active to keep compatibility with current code base
	stdLog.SetOutput(logFileObj)
	stdLog.SetPrefix("LOG: ")
	stdLog.SetFlags(stdLog.Ldate)

	log.Logger = zeroLog.New(logFileObj).With().Timestamp().Logger()

	viper.Set("k1-paths.logs-dir", logsFolder)
	viper.Set("k1-paths.log-file", logfile)
	viper.Set("k1-paths.log-file-name", logfileName)

	err = viper.WriteConfig()
	if err != nil {
		stdLog.Panicf("unable to set log-file-location, error is: %s", err)
	}

	canRunBubbleTea = common.Debug
	_, present := os.LookupEnv("KUBEFIRST_DISABLE_BUBBLETEA")

	// disable bubbletea if env var is set, slow in vscode (TODO: need to figure out why)
	if present {
		canRunBubbleTea = false
	}

	if canRunBubbleTea {
		progress.InitializeProgressTerminal()

		go func() {
			cmd.Execute()
		}()

		progress.Progress.Run()
	} else {
		cmd.Execute()
	}

}
