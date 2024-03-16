package validator

import (
	"net/http"
	"testing"

	"github.com/go-playground/validator"
	"github.com/stretchr/testify/assert"
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
				validator: validator.New(),
			}
			err := gv.Validate(tt.args.i)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, 400)
				assert.Contains(t, err.Error(), "received invalid request body")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
