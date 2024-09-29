package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	DataciteRecordArchivePath  string `json:"dataciteRecordArchivePath"`
	DataciteRecordWorkerNumber int    `json:"dataciteRecordWorkerNumber"`
	OutputDir                  string `json:"outputDir"`
	SizeOfPayloadChunk         int    `json:"sizeOfPayloadChunk"`
}

// returns default configuration on error condition
func FromFile() (Config, error) {
	// default values, if not set in file
	config := Config{
		DataciteRecordArchivePath:  "/tmp/data.tar.gz",
		DataciteRecordWorkerNumber: 4,
		OutputDir:                  "/tmp/out",
		SizeOfPayloadChunk:         4096,
	}

	if len(os.Args) < 2 {
		return config, fmt.Errorf("Config file path was not specified on first CLI param")
	}
	filepath := os.Args[1]
	configFile, err := os.Open(filepath)

	if err != nil {
		return config, err
	}

	defer configFile.Close()

	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func HashConfig(config Config) (string, error) {
	configBytesGolang, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	// Prepare String-representation of Config: we need the same result as here:
	// https://github.com/tgweber/mlmetacode/blob/491a52972c515be7658325f1503db005fde18d01/code/util/util.py#L88C27-L88C82
	// Python adds additional whitespace, so we mimick that behaviour
	replacer := strings.NewReplacer("\":", "\": ", ",\"", ", \"")
	configStringPython := replacer.Replace(string(configBytesGolang))

	hasher := sha256.New()
	hasher.Write([]byte(configStringPython))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
