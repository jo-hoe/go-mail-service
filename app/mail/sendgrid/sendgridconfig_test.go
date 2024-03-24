package sendgrid

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jo-hoe/go-mail-service/app/mail"
)

func Test_Integration_createConfig(t *testing.T) {
	os.Setenv(defaultAddressEnvKey, "default@example.com")
	defer os.Unsetenv(defaultAddressEnvKey)
	os.Setenv(defaultNameEnvKey, "default-name")
	defer os.Unsetenv(defaultNameEnvKey)


	content := "super-secret-api-key"	
	rootDirectory, fileName := setupIntegrationTestFile(t, content)
	filePath := filepath.Join(rootDirectory, fileName)
	// order of defers matter as file need to be closed
	// before we can delete the folder
	defer func() {
		err := os.RemoveAll(rootDirectory)
		if err != nil {
			t.Errorf("could not delete file '%+v'", err)
		}
	}()

	type args struct {
		mailAttributes mail.MailAttributes
	}
	tests := []struct {
		name    string
		args    args
		want    *SendGridConfig
		wantErr bool
	}{
		{
			name: "positive case overwrite default",
			args: args{
				mailAttributes: mail.MailAttributes{
					From:     "from@example.com",
					FromName: "FromNameExample",
				},
			},
			want: &SendGridConfig{
				APIKey:        "super-secret-api-key",
				OriginAddress: "from@example.com",
				OriginName:    "FromNameExample",
			},
			wantErr: false,
		},
		{
			name: "positive case use defaults",
			args: args{
				mailAttributes: mail.MailAttributes{},
			},
			want: &SendGridConfig{
				APIKey:        "super-secret-api-key",
				OriginAddress: "default@example.com",
				OriginName:    "default-name",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {			
			file, err := os.Open(filePath)
			if err != nil {
				t.Error("could not open file")
			}
			defer func() {
				err := file.Close()
				if err != nil {
					t.Errorf("could not delete file '%+v'", err)
				}
			}()

			got, err := createConfig(tt.args.mailAttributes, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("createConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createConfig(t *testing.T) {
	os.Setenv(defaultAddressEnvKey, "default@example.com")
	defer os.Unsetenv(defaultAddressEnvKey)
	os.Setenv(defaultNameEnvKey, "default-name")
	defer os.Unsetenv(defaultNameEnvKey)

	type args struct {
		mailAttributes mail.MailAttributes
	}
	tests := []struct {
		name    string
		args    args
		want    *SendGridConfig
		wantErr bool
	}{
		{
			name: "missing file",
			args: args{
				mailAttributes: mail.MailAttributes{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createConfig(tt.args.mailAttributes, AlwaysErrorReader{})
			if (err == nil) {
				t.Errorf("createConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createConfig() = %v, want %v", got, tt.want)
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

type AlwaysErrorReader struct{}

func (a AlwaysErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("always error")
}
