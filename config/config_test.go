package config

import (
	"testing"
)

func TestInitConfiguration(t *testing.T) {
	// Happy path
	config, err := InitConfiguration("../config-example.toml")
	if err != nil {
		t.Fatal(err)
	}
	if config == nil {
		t.Fatal("expected config, got nil")
	}

	// File does not exist
	_, err = InitConfiguration("does_not_exist.toml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Invalid YAML
	_, err = InitConfiguration("invalid.toml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
