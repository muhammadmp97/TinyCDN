package utils

import (
	"testing"
)

func TestGetExtension(t *testing.T) {
	t.Run("GetExtension() ingores query strings", func(t *testing.T) {
		if GetExtension("https://example.com/file.jpg?d=x.css") != ".jpg" {
			t.Error("GetExtension didn't ignore the query string.")
		}
	})

	t.Run("GetExtension() returns an empty string on failure", func(t *testing.T) {
		if GetExtension("https://example.com/file") != "" {
			t.Error("GetExtension didn't return an empty string.")
		}
	})
}
