package main

import "testing"

func TestLoadConfig(t *testing.T) {
	_, err := loadConfig()
	if err != nil {
		t.Fatal("loadConfig() error", err)
	}
}
