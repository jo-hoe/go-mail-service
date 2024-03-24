package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSecretFileService_Get(t *testing.T) {
	type fields struct {
		handle io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Get secret from file",
			fields: fields{
				handle: bytes.NewReader([]byte("secret")),
			},
			want:    "secret",
			wantErr: false,
		},
		{
			name: "Get secret from empty file",
			fields: fields{
				handle: bytes.NewReader([]byte("")),
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Get secret error throwing reader",
			fields: fields{
				handle: AlwaysErrorReader{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSecretFileService()
			got, err := s.Get(tt.fields.handle)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretFileService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SecretFileService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setupIntegrationTestFile(t *testing.T, content string) (rootDirectory string, fileName string) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	if err != nil {
		t.Error("could not create folder")
	}
	file, err := os.CreateTemp(rootDirectory, "testFile")
	if err != nil {
		t.Error("could not create file")
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		t.Errorf("could not write to file %+v", err)
	}
	fileName = filepath.Base(file.Name())
	if err != nil {
		t.Errorf("could not close file %+v", err)
	}

	return rootDirectory, fileName
}

func Test_Integration_SecretFileService_Get(t *testing.T) {
	content := "secret\ncontent"
	rootDirectory, fileName := setupIntegrationTestFile(t, content)
	filePath := filepath.Join(rootDirectory, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		t.Error("could not open file")
	}

	// order of defers matter as file need to be closed
	// before we can delete the folder
	defer func() {
		err := os.RemoveAll(rootDirectory)
		if err != nil {
			t.Errorf("could not delete file '%+v'", err)
		}
	}()
	defer file.Close()

	type fields struct {
		handle io.Reader
	}

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Get secret from file",
			fields: fields{
				handle: file,
			},
			want:    content,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSecretFileService()
			got, err := s.Get(tt.fields.handle)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretFileService.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SecretFileService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

type AlwaysErrorReader struct{}

func (a AlwaysErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("always error")
}
