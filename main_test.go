package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFailOnErrorWithNilError(t *testing.T) {
	FailOnError(nil, "should not fail")
}

func TestLoadEnvFileIgnoresMissingFile(t *testing.T) {
	if err := loadEnvFile(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("expected missing env file to be ignored, got %v", err)
	}
}

func TestLoadEnvFileReturnsParseAndReadErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.env")
	if err := os.WriteFile(path, []byte("BAD@KEY=value\n"), 0o600); err != nil {
		t.Fatalf("write malformed env file: %v", err)
	}

	err := loadEnvFile(path)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("expected parse error containing path, got %v", err)
	}

	if err := loadEnvFile(t.TempDir()); err == nil {
		t.Fatal("expected directory read error")
	}
}
