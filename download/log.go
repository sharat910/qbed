package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func createDirs(filePath string) {
	//Creating directories
	directorystring := filepath.Dir(filePath)
	err := os.MkdirAll(directorystring, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create directories")
	}
}

func getLevel(l string) zerolog.Level {
	switch l {
	case "info":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		fmt.Println("unknown level string: setting logging level to info")
		return zerolog.InfoLevel
	}
}

func setupLogging(l string) {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.999"
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	logpath := filepath.Join("logs", fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	createDirs(logpath)
	file, err := os.Create(logpath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create log file")
	}
	multi := zerolog.MultiLevelWriter(consoleWriter, file)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(getLevel(l))
	log.Info().Str("Level", l).Msg("Logging setup done")
}
