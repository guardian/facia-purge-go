package main

import (
	"testing"
)

func TestExtractFront(t *testing.T) {
	path := "DEV/frontsapi/pressed/live/au/sport/fapi/pressed.v2.json"
	expected := "au/sport"
	actual := extractFront(path)
	if actual != expected {
		t.Errorf("Failed to extract front, expected %s but got %s", expected, actual)
	}
}
