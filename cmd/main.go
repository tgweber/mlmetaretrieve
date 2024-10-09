package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"github.com/tgweber/mlmetaretrieve/internal/clog"
	"github.com/tgweber/mlmetaretrieve/internal/config"
	"github.com/tgweber/mlmetaretrieve/internal/datacite"
)

type StatusMessage struct {
	Worker      string
	Message     string
	Error       error
	RecordsSeen int
	RecordsUsed int
}

type Statistics struct {
	RecordsSeen int
	RecordsUsed int
}

func main() {
	config, err := config.FromFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Logging
	logger := clog.SetupLogger(config)
	logger.Info("Configuration: %s", config)

	// Output setup
	err = os.MkdirAll(config.OutputDir, os.ModePerm)
	if err != nil {
		logger.Error("Cannot create output dir logfile", "error", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	statusMessageCh := make(chan StatusMessage)
	dataciteRecordCh := readFromArchive(config, statusMessageCh)
	processDataciteRecords(&wg, config, dataciteRecordCh, statusMessageCh)

	stats := Statistics{}
	regexForWorkerFinished := regexp.MustCompile("Worker [0-9]+ finished")
	for msg := range statusMessageCh {
		if msg.Error != nil {
			logger.Error("ERROR", "error", msg.Error.Error(), "worker", msg.Worker)
			continue
		}
		if regexForWorkerFinished.Match([]byte(msg.Message)) {
			stats.RecordsSeen += msg.RecordsSeen
			stats.RecordsUsed += msg.RecordsUsed
		}
		logger.Info("Event", "worker", msg.Worker, "message", msg.Message)
	}
	logger.Info("Finished processing", "recordsSeen", stats.RecordsSeen, "recordsUsed", stats.RecordsUsed)
}

func processDataciteRecords(wg *sync.WaitGroup, config config.Config, in <-chan []byte, out chan<- StatusMessage) {
	for i := 0; i < config.DataciteRecordWorkerNumber; i++ {
		worker := strconv.Itoa(i)
		wg.Add(1)
		go func() {
			recordsSeen := 0
			recordsUsed := 0
			useableRecords := datacite.DatacitePayload{
				Documents: []datacite.DataciteRecord{},
			}
			for payload := range in {
				recordsSeen += 1
				dataciteRecord := datacite.DataciteRecord{}

				err := json.Unmarshal(payload, &dataciteRecord)
				if err != nil {
					out <- StatusMessage{
						Message: "Error",
						Error:   fmt.Errorf("Could not create datacite payload: %v payload: %s", err, string(payload)),
						Worker:  worker,
					}
				}
				if dataciteRecord.IsUseable() {
					recordsUsed += 1
					useableRecords.Add(dataciteRecord)
				}
				if len(useableRecords.Documents) == config.SizeOfPayloadChunk {
					outputFile := filepath.Join(config.OutputDir, fmt.Sprintf("%s.json", uuid.New().String()))
					err := useableRecords.Flush(outputFile)
					if err != nil {
						out <- StatusMessage{
							Message: "Error",
							Error:   fmt.Errorf("Could not flush datacite records: %v", err),
							Worker:  worker,
						}
					}
				}
			}
			if len(useableRecords.Documents) > 0 {
				outputFile := filepath.Join(config.OutputDir, fmt.Sprintf("%s.json", uuid.New().String()))
				err := useableRecords.Flush(outputFile)
				if err != nil {
					out <- StatusMessage{
						Message: "Error",
						Error:   fmt.Errorf("Could not flush datacite records: %v", err),
						Worker:  worker,
					}
				}
			}
			out <- StatusMessage{
				Message:     fmt.Sprintf("Worker %s finished.", worker),
				RecordsSeen: recordsSeen,
				RecordsUsed: recordsUsed,
				Worker:      worker,
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func readFromArchive(config config.Config, messageCh chan<- StatusMessage) <-chan []byte {
	out := make(chan []byte)
	go func() {
		f, err := os.Open(config.DataciteRecordArchivePath)
		if err != nil {
			messageCh <- StatusMessage{
				Message: "Error",
				Error:   err,
				Worker:  "readFromArchive",
			}
			close(out)
			return
		}

		defer func() {
			err := f.Close()
			if err != nil {
				messageCh <- StatusMessage{
					Message: "Error",
					Error:   err,
					Worker:  "readFromArchive",
				}
			}
		}()

		gzf, err := gzip.NewReader(f)
		if err != nil {
			messageCh <- StatusMessage{
				Message: "Error",
				Error:   err,
				Worker:  "readFromArchive",
			}
			return
		}

		tarReader := tar.NewReader(gzf)

		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				messageCh <- StatusMessage{
					Message: "Error",
					Error:   err,
					Worker:  "readFromArchive",
				}
				return

			}

			if header.Typeflag == tar.TypeReg {
				messageCh <- StatusMessage{
					Message: fmt.Sprintf("Processing %s (size %d)", header.Name, header.Size),
					Worker:  "readFromArchive",
				}
				scanner := bufio.NewScanner(tarReader)
				scanner.Buffer(make([]byte, 8192*1024), 4096*1024)
				for scanner.Scan() {
					out <- []byte(scanner.Text())
				}

				if scanner.Err() != nil {
					messageCh <- StatusMessage{
						Message: "Error",
						Error:   scanner.Err(),
						Worker:  "readFromArchive",
					}
				}
			}

		}
		close(out)
	}()
	return out
}
