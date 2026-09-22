package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIValidationAndFailedExit(t *testing.T) {
	var output, errors bytes.Buffer
	if code := run([]string{"6", "Manual", "3000010000000010"}, &output, &errors); code != 2 {
		t.Fatalf("invalid target exit = %d, want 2", code)
	}
	if !strings.Contains(errors.String(), "usage:") {
		t.Fatalf("invalid target error = %q", errors.String())
	}
	errors.Reset()
	if code := run([]string{"1", "Manual"}, &output, &errors); code != 2 {
		t.Fatalf("missing account exit = %d, want 2", code)
	}
	errors.Reset()
	if code := run([]string{"1", "Other", "3000010000000010"}, &output, &errors); code != 2 {
		t.Fatalf("invalid change type exit = %d, want 2", code)
	}
	if !strings.Contains(errors.String(), "Manual or Automatic") {
		t.Fatalf("invalid change type error = %q", errors.String())
	}

	t.Setenv("FINCLOUD_BASE_URL", "http://fincloud.test")
	t.Setenv("FINCLOUD_USERNAME", "user")
	t.Setenv("FINCLOUD_PASSWORD", "password")
	t.Setenv("FINCLOUD_ROLE_ID", "role")
	output.Reset()
	errors.Reset()
	if code := run([]string{"1", "manual", "123"}, &output, &errors); code != 1 {
		t.Fatalf("invalid account exit = %d, want 1", code)
	}
	if !strings.Contains(output.String(), "change Manual | FAILED") || !strings.Contains(output.String(), "Failed:    1") {
		t.Fatalf("invalid account output = %q", output.String())
	}
}

func TestCLILoadsEnvFile(t *testing.T) {
	for _, name := range []string{"FINCLOUD_BASE_URL", "FINCLOUD_USERNAME", "FINCLOUD_PASSWORD", "FINCLOUD_ROLE_ID"} {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
	t.Setenv("FINCLOUD_USERNAME", "shell-user")
	directory := t.TempDir()
	contents := "FINCLOUD_BASE_URL=http://fincloud.test\nFINCLOUD_USERNAME=file-user\nFINCLOUD_PASSWORD=password\nFINCLOUD_ROLE_ID=role\n"
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(previous) })
	var output, errors bytes.Buffer
	if code := run([]string{"1", "automatic", "123"}, &output, &errors); code != 1 {
		t.Fatalf(".env run exit = %d, error = %q", code, errors.String())
	}
	if !strings.Contains(output.String(), "change Automatic | FAILED") {
		t.Fatalf("canonical change type missing from output: %q", output.String())
	}
	if got := os.Getenv("FINCLOUD_USERNAME"); got != "shell-user" {
		t.Fatalf("environment precedence = %q, want shell-user", got)
	}
	if got := os.Getenv("FINCLOUD_PASSWORD"); got != "password" {
		t.Fatalf(".env password was not loaded")
	}
}
