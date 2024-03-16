package validation

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator"
)

type User struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestGenericValidator_Validate(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid user",
			args: args{
				i: &User{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid user - missing name",
			args: args{
				i: &User{
					Email: "john.doe@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid user - invalid email",
			args: args{
				i: &User{
					Name:  "John Doe",
					Email: "john.doe",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gv := &GenericValidator{
				Validator: validator.New(),
			}
			err := gv.Validate(tt.args.i)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
				if http.StatusBadRequest != 400 {
					t.Errorf("Validate() error status = %v, want %v", http.StatusBadRequest, 400)
				}
				if !strings.Contains(err.Error(), "received invalid request body") {
					t.Errorf("Validate() error message = %v, want %v", err.Error(), "received invalid request body")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
