package config

import (
	"os"
	"strings"
	"testing"
)

func TestEnvService_Get(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "existing key",
			args:    args{key: "SOME_KEY", value: "some-value"},
			want:    "some-value",
			wantErr: false,
		},
		{
			name:    "non-existing key",
			args:    args{},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.key != "" {
				os.Setenv(tt.args.key, tt.want)
				defer os.Unsetenv(tt.args.key)
			}

			s := NewEnvService()
			got, err := s.Get(tt.args.key)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EnvService.Get() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("EnvService.Get() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if strings.Compare(tt.want, got) != 0 {
				t.Errorf("EnvService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
