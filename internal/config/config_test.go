package config

import (
	"testing"
)

func TestHashConfig(t *testing.T) {
	want := "f7301e69a40465772339f8548d356e8087b3b27df3b2f954acf9b565d939fb44"
	config := Config{
		DataciteRecordArchivePath:  "/tmp/small.tar.gz",
		DataciteRecordWorkerNumber: 15,
		OutputDir:                  "/tmp/small",
		SizeOfPayloadChunk:         32768,
	}
	have, _ := HashConfig(config)
	if want != have {
		t.Fatalf(`HashConfig returns %s, %s is wanted`, have, want)
	}
}
