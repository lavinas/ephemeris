package pkg

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestIsEmpty",
			args: args{
				i: "",
			},
			want: true,
		},
		{
			name: "TestIsEmpty",
			args: args{
				i: nil,
			},
			want: true,
		},
		{
			name: "TestIsEmpty",
			args: args{
				i: 0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.args.i); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
