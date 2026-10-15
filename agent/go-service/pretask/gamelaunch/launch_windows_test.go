package gamelaunch

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

type childReceipt struct {
	Parent int
	Dir    string
	Args   []string
}

func TestNativeParent(t *testing.T) {
	if os.Getenv("MAAEND_LAUNCH_RECEIPT") != "" {
		_, _ = io.Copy(io.Discard, os.Stdin)
	}
}

func TestNativeChild(t *testing.T) {
	path := os.Getenv("MAAEND_LAUNCH_RECEIPT")
	if path == "" {
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(childReceipt{os.Getppid(), dir, os.Args[3:]})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

// 用无副作用的测试进程验证实际父 PID、工作目录和含空格/非 ASCII 参数，不启动游戏。
func TestNativeCreateGame(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	receipt := filepath.Join(t.TempDir(), "receipt.json")
	t.Setenv("MAAEND_LAUNCH_RECEIPT", receipt)
	parent := exec.Command(exe, "-test.run=^TestNativeParent$")
	stdin, err := parent.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := parent.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { stdin.Close(); _ = parent.Wait() }()
	handle, err := windows.OpenProcess(windows.PROCESS_CREATE_PROCESS|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(parent.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	process, err := createGame(exe, `-test.run=^TestNativeChild$ -- "含空格 value" "-channel=7"`, handle)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(process.Process)
	windows.CloseHandle(process.Thread)
	state, err := windows.WaitForSingleObject(process.Process, 10000)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		_ = windows.TerminateProcess(process.Process, 1)
		t.Fatalf("test child did not exit: %d, %v", state, err)
	}
	data, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var got childReceipt
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Parent != parent.Process.Pid || !strings.EqualFold(got.Dir, filepath.Dir(exe)) || !reflect.DeepEqual(got.Args, []string{"含空格 value", "-channel=7"}) {
		t.Fatalf("unexpected native launch: %+v; expected parent %d", got, parent.Process.Pid)
	}
}

func TestRejectUntrustedSignatures(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyLauncher(exe); err == nil {
		t.Fatal("unsigned test executable accepted")
	}
	if _, err := verifyLauncher(filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")); err == nil {
		t.Fatal("valid signature from a different publisher accepted")
	}
}

func TestOfficialLauncherSignature(t *testing.T) {
	path := os.Getenv("MAAEND_TEST_LAUNCHER")
	if path == "" {
		t.Skip("set MAAEND_TEST_LAUNCHER to verify an installed official launcher")
	}
	if fingerprint, err := verifyLauncher(path); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("verified official launcher signer: %s", fingerprint)
	}
}

func TestRunningOfficialParent(t *testing.T) {
	if os.Getenv("MAAEND_TEST_LAUNCHER") == "" {
		t.Skip("installed official launcher check is opt-in")
	}
	handle, err := officialParent()
	if err != nil {
		t.Fatal(err)
	}
	if handle == 0 {
		t.Skip("official launcher is not running")
	}
	defer windows.CloseHandle(handle)
	pid, err := windows.GetProcessId(handle)
	if err != nil || pid == 0 {
		t.Fatalf("invalid verified parent handle: %d, %v", pid, err)
	}
	t.Logf("verified running parent PID %d", pid)
}

func TestAlreadyRunning(t *testing.T) {
	ids, err := processIDs("Endfield.exe")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Skip("no running game; do not launch one in tests")
	}
	if !Run([]string{"invalid config is ignored when game already runs"}) {
		t.Fatal("running game should skip all launch configuration")
	}
	if ready, err := gameWindowReady(ids[0]); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("existing game visible Unity window: %v", ready)
	}
}
