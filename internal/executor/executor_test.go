package executor

import (
	"bytes"
	"io"
	"os/exec"
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	e := New()

	if e == nil {
		t.Fatal("expected non-nil executor")
	}
	if e.Stdin == nil {
		t.Error("expected non-nil Stdin")
	}
	if e.Stdout == nil {
		t.Error("expected non-nil Stdout")
	}
	if e.Stderr == nil {
		t.Error("expected non-nil Stderr")
	}
}

func TestExecutor_RunSimple(t *testing.T) {
	e := New()

	var stdout bytes.Buffer
	e.Stdout = &stdout

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo", "hello")
	} else {
		cmd = exec.Command("echo", "hello")
	}

	exitCode, err := e.RunSimple(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestExecutor_RunSimple_NonZeroExit(t *testing.T) {
	e := New()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "exit", "1")
	} else {
		cmd = exec.Command("sh", "-c", "exit 1")
	}

	exitCode, err := e.RunSimple(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestExecutor_RunSimple_CommandNotFound(t *testing.T) {
	e := New()

	cmd := exec.Command("nonexistent-command-12345")

	_, err := e.RunSimple(cmd)
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestGracefulShutdownTimeout(t *testing.T) {
	if GracefulShutdownTimeout.Seconds() != 5 {
		t.Errorf("expected 5 second timeout, got %v", GracefulShutdownTimeout)
	}
}

func TestCancelableReader_NormalRead(t *testing.T) {
	data := []byte("hello world")
	inner := bytes.NewReader(data)
	cr := &cancelableReader{r: inner}

	buf := make([]byte, len(data))
	n, err := cr.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes, got %d", len(data), n)
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("expected %q, got %q", data, buf)
	}
}

func TestCancelableReader_CancelBeforeRead(t *testing.T) {
	inner := bytes.NewReader([]byte("hello"))
	cr := &cancelableReader{r: inner}

	// Cancel before reading
	cr.Cancel()

	buf := make([]byte, 10)
	n, err := cr.Read(buf)
	if err != io.EOF {
		t.Errorf("expected io.EOF after cancel, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 bytes after cancel, got %d", n)
	}
}

func TestCancelableReader_MultipleCancelSafe(t *testing.T) {
	inner := bytes.NewReader([]byte("hello"))
	cr := &cancelableReader{r: inner}

	// Multiple cancels should be safe
	cr.Cancel()
	cr.Cancel()
	cr.Cancel()

	buf := make([]byte, 10)
	_, err := cr.Read(buf)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}
