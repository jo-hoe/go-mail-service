package sendgrid

import (
	"os"
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
	os.Setenv(apiEnvKey, content)
	defer os.Unsetenv(apiEnvKey)

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
			got, err := createConfig(tt.args.mailAttributes)
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
			got, err := createConfig(tt.args.mailAttributes)
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