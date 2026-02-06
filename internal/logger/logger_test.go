package logger

import (
	"bytes"
	"os"
	"testing"
)

func TestInfo_Success_Warn_Error_NoPanic(t *testing.T) {
	// Redirect stdout so we don't spam the test output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	Info("TAG", "message")
	Success("TAG", "message")
	Warn("TAG", "message")
	Error("TAG", "message")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	// Just ensure we didn't panic; output is environment-dependent (colors, etc.)
}

func TestBanner_NoPanic(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	Banner("v1.0.0")
	Banner("")

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
}

func TestSectionAndStats_NoPanic(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()
	Section("Test")
	Stats("key", 42)
	w.Close()
}
