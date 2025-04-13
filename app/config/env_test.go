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
                if err := os.Setenv(tt.args.key, tt.want); err != nil {
                    t.Fatalf("Failed to set environment variable: %v", err)
                }
                defer func() {
                    if err := os.Unsetenv(tt.args.key); err != nil {
                        t.Fatalf("Failed to unset environment variable: %v", err)
                    }
                }()
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