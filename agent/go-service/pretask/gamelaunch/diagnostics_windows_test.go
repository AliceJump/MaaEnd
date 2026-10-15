package gamelaunch

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestSDKDiagnosticsOnlyNewChecks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sdk.log")
	old := "Parent process check success:old\n"
	newRecords := "Parent process fingerprint not matched:new\nOnU8V2InitCallback status :-1\n"
	if err := os.WriteFile(path, []byte(old+newRecords+"unrelated account data\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&output)
	defer func() { log.Logger = previousLogger }()
	sdkLogs{{path: path, offset: int64(len(old))}}.report()
	var event struct {
		NewCheckRecords []string `json:"new_check_records"`
	}
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	if len(event.NewCheckRecords) != 2 || event.NewCheckRecords[0] != "Parent process fingerprint not matched:new" || event.NewCheckRecords[1] != "OnU8V2InitCallback status :-1" {
		t.Fatalf("unexpected SDK diagnostics: %s", output.String())
	}
}
