package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type CSVRecord interface {
	GetType() string
	GetHeader() []string
	GetRow() []string
}

type AutoCSVRecord struct {
	rec interface{}
}

func (a AutoCSVRecord) GetType() string {
	return AutoName(a.rec)
}

func (a AutoCSVRecord) GetHeader() []string {
	return AutoHeader(a.rec)
}

func (a AutoCSVRecord) GetRow() []string {
	return AutoRow(a.rec)
}

func AutoName(item interface{}) string {
	t := reflect.TypeOf(item)
	return strings.ToLower(t.Name())
}

func AutoHeader(item interface{}) []string {
	s := reflect.ValueOf(item)
	var header []string
	for i := 0; i < s.NumField(); i++ {
		header = append(header, s.Type().Field(i).Name)
	}
	return header
}

func AutoRow(item interface{}) []string {
	s := reflect.ValueOf(item)
	var row []string
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			row = append(row, strconv.FormatInt(f.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			row = append(row, strconv.FormatUint(f.Uint(), 10))
		case reflect.Float32, reflect.Float64:
			row = append(row, strconv.FormatFloat(f.Float(), 'f', -1, 64))
		case reflect.String:
			row = append(row, f.String())
		case reflect.Struct:
			if f.Type().Name() == "Time" {
				row = append(row, f.Interface().(time.Time).Format("2006-01-02 15:04:05.999999"))
			}
		}
	}
	return row
}

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
	log.Info().Str("path", rootPath).Msg("csv dumper ready")
	for c := range csvrecords {
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
