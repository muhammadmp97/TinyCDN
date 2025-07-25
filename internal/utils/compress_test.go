package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"testing"
)

func TestIsCompressible(t *testing.T) {
	t.Run("Javascript files are compressible", func(t *testing.T) {
		if !IsCompressible("https://example.com/app.js") {
			t.Error("Javascript files should be compressible!")
		}
	})

	t.Run("XYZ files are not compressible", func(t *testing.T) {
		if IsCompressible("https://example.com/file.xyz") {
			t.Error("Not included file formats should not be compressible!")
		}
	})
}

func TestCompress(t *testing.T) {
	t.Run("Compress() works properly", func(t *testing.T) {
		original := []byte("body { background-color: red; }")
		compressed := []byte(Compress(&original))

		reader, _ := gzip.NewReader(bytes.NewReader(compressed))
		defer reader.Close()

		decompressed, _ := ioutil.ReadAll(reader)
		if !bytes.Equal(decompressed, original) {
			t.Errorf("Decompressed data does not match original!\nActual: %s\nExpected: %s", decompressed, original)
		}
	})
}
