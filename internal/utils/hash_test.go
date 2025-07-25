package utils

import (
	"testing"
)

func TestXXHash(t *testing.T) {
	t.Run("XXHash() works properly", func(t *testing.T) {
		actual := XXHash("https://example.com/style.css:gzip")
		expected := "9fed30d501abe399"

		if actual != expected {
			t.Errorf("The output doesn't match what's expected!\nActual: %s\nExpected: %s", actual, expected)
		}
	})
}
