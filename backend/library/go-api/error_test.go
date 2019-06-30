package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	type args struct {
		code    int
		message string
	}
	tests := []struct {
		name string
		args args
		want Error
	}{
		{
			"standard case",
			args{1, "something went wrong"},
			Error{
				Code:    int32(1),
				Message: "something went wrong",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := NewError(tt.args.code, tt.args.message)
			assert.Equal(t, tt.want, got)
		})
	}
}