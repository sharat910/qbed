package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
)

type FileAndWriter struct {
	File   *os.File
	Writer *csv.Writer
}

func GetFileWriter(filePath string) FileAndWriter {
	createDirs(filePath)
	f := createFile(filePath)
	return FileAndWriter{
		File:   f,
		Writer: csv.NewWriter(f),
	}
}

func createDirs(filePath string) {
	//Creating directories
	directorystring := filepath.Dir(filePath)
	err := os.MkdirAll(directorystring, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create directories")
	}
}

func createFile(filePath string) *os.File {
	//Creating file
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create file")
	}
	return file
}

func CSVDumper(rootPath string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	FWMap := make(map[string]FileAndWriter, 2)
	var fw FileAndWriter
	var found bool

	for c := range stats {
		//fmt.Println(c.GetType(), c.GetRow())
		fw, found = FWMap[c.GetType()]
		if !found {
			fw = GetFileWriter(filepath.Join(rootPath, c.GetType()+".csv"))
			err := fw.Writer.Write(c.GetHeader())
			if err != nil {
				log.Fatal().Err(err).Msg("unable to write header")
			}
			FWMap[c.GetType()] = fw
		}
		err := fw.Writer.Write(c.GetRow())
		if err != nil {
			log.Fatal().Err(err).Msg("unable to write row")
		}
	}
	for _, fw := range FWMap {
		fw.Writer.Flush()
		err := fw.File.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("unable to close file")
		}
	}
	log.Info().Msg("csv files flushed")
}
