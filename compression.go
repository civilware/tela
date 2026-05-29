package tela

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"compress/gzip"

	"github.com/andybalholm/brotli"
)

const (
	COMPRESSION_GZIP   = ".gz"
	COMPRESSION_BROTLI = ".br"
)

// Compression formats this package will use
var compressionFormats = []string{
	COMPRESSION_GZIP,
	COMPRESSION_BROTLI,
}

// TrimCompressedExt removes known compression extensions from the filename
func TrimCompressedExt(fileName string) string {
	if fileName == "" {
		return fileName
	}

	for _, comp := range compressionFormats {
		fileName = strings.TrimSuffix(fileName, comp)
	}

	return fileName
}

// IsCompressedExt returns true if ext is a valid TELA compression format
func IsCompressedExt(ext string) bool {
	if ext == "" {
		return false
	}

	for _, comp := range compressionFormats {
		if ext == comp {
			return true
		}
	}

	return false
}

// Decompress TELA data using the given compression format, if compression is "" result will return the original data
func Decompress(data []byte, compression string) (result []byte, err error) {
	switch compression {
	case COMPRESSION_GZIP:
		result, err = decompressGzip(data)
		if err != nil {
			return
		}
	case COMPRESSION_BROTLI:
		result, err = decompressBrotli(data)
		if err != nil {
			return
		}
	case "":
		result = data
	default:
		err = fmt.Errorf("unknown decompression format %q", compression)
	}

	return
}

// Compress TELA data using the given compression format
func Compress(data []byte, compression string) (result string, err error) {
	switch compression {
	case COMPRESSION_GZIP:
		result, err = compressGzip(data)
		if err != nil {
			return
		}
	case COMPRESSION_BROTLI:
		result, err = compressBrotli(data)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("unknown compression format %q", compression)
	}

	return
}

// Compress data as gzip then encode it in base64 and return the result
func compressGzip(data []byte) (result string, err error) {
	var buf bytes.Buffer
	var gz *gzip.Writer
	gz, err = gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return
	}
	defer gz.Close()

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	// Ensure all data is written to the buffer
	err = gz.Close()
	if err != nil {
		return
	}

	result = base64.StdEncoding.EncodeToString(buf.Bytes())

	return
}

// Decompress base64 encoded gzip data and return the result
func decompressGzip(data []byte) (result []byte, err error) {
	var decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return
	}

	var gz *gzip.Reader
	gz, err = gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return
	}
	defer gz.Close()

	var decompressed []byte
	decompressed, err = io.ReadAll(gz)
	if err != nil {
		return
	}

	result = decompressed

	return
}

// Compress data as brotli then encode it in base64 and return the result
func compressBrotli(data []byte) (result string, err error) {
	var buf bytes.Buffer
	bw := brotli.NewWriterLevel(&buf, brotli.BestCompression)

	_, err = bw.Write(data)
	if err != nil {
		return
	}

	// Ensure all data is written to the buffer
	err = bw.Close()
	if err != nil {
		return
	}

	result = base64.StdEncoding.EncodeToString(buf.Bytes())

	return
}

// Decompress base64 encoded brotli data and return the result
func decompressBrotli(data []byte) (result []byte, err error) {
	var decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return
	}

	br := brotli.NewReader(bytes.NewReader(decoded))

	var decompressed []byte
	decompressed, err = io.ReadAll(br)
	if err != nil {
		return
	}

	result = decompressed

	return
}
