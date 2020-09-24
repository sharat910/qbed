package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func Server(addr string) {
	log.Info().Str("addr", addr).Msg("starting http file sever")
	http.HandleFunc("/", HandleLayman)
	http.HandleFunc("/download", Handle)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start http server")
	}
}

type ByteGen struct {
	TotalBytes int64
}

func (b ByteGen) Read(p []byte) (n int, err error) {
	return len(p), nil
}

func (b ByteGen) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekEnd {
		return b.TotalBytes, nil
	}
	return 0, nil
}

func Handle(writer http.ResponseWriter, request *http.Request) {
	fileSize, err := strconv.Atoi(request.URL.Query().Get("size"))
	if err != nil {
		//Get not set, send a 400 bad request
		http.Error(writer, "Error parsing size: "+fmt.Sprint(err), 400)
		return
	}

	log.Info().Int("size", fileSize).Str("range", request.Header.Get("Range")).Msg("serving download")
	//// Create the file content
	//file := make([]byte, fileSize)
	//readSeeker := bytes.NewReader(file)
	readSeeker := ByteGen{TotalBytes: int64(fileSize)}
	http.ServeContent(writer, request, "d.bin", time.Now(), readSeeker)
}

func HandleLayman(writer http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	fileSize, err := strconv.Atoi(request.URL.Query().Get("size"))
	if err != nil {
		//Get not set, send a 400 bad request
		http.Error(writer, "Error parsing size: "+fmt.Sprint(err), 400)
		return
	}

	// Create the file
	file := make([]byte, fileSize)

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%dbytes", fileSize))
	writer.Header().Set("Content-Type", "application/octet-stream")

	_, err = writer.Write(file)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to write content")
	}

	return
}
