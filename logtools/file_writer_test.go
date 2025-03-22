package logtools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileWriter_Close(t *testing.T) {
	file, err := os.CreateTemp("", "file-writer.test")

	errMsg := "expected no error when creating temp file, got %v"
	require.Nilf(t, err, errMsg, err)

	name := file.Name()
	defer os.Remove(name)

	fw, err := NewFileWriter(name)

	errMsg = "expected no error when creating file writer file, got %v"
	require.Nilf(t, err, errMsg, err)

	mockPayload := []byte(`{"msg":"test"}`)
	_, err = fw.Write(mockPayload)

	errMsg = "expected no error when writing log, got %v"
	require.Nilf(t, err, errMsg, err)

	err = fw.Close()
	errMsg = "expected no error when closing file, got %v"
	require.Nilf(t, err, errMsg, err)

	payload, err := os.ReadFile(name)

	errMsg = "expected no error when reading file, got %v"
	require.Nilf(t, err, errMsg, err)

	errMsg = "filed content must be equal to mock payload"
	require.Equal(t, mockPayload, payload, errMsg)
}
