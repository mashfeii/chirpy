package stringshelpers_test

import (
	"testing"

	stringshelpers "github.com/mashfeii/chirpy/pkg/strings_helpers"
)

func TestCleanString(t *testing.T) {
	type args struct {
		initial string
		stop    []string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Replace hello",
			args: args{
				initial: "hello world",
				stop:    []string{"hello"},
			},
			want: "**** world",
		},
		{
			name: "Replace world",
			args: args{
				initial: "hello world",
				stop:    []string{"world"},
			},
			want: "hello ****",
		},
		{
			name: "Unchanged",
			args: args{
				initial: "hello world",
				stop:    []string{"hi", "bye"},
			},
			want: "hello world",
		},
		{
			name: "Exact match",
			args: args{
				initial: "hello? world!",
				stop:    []string{"hello", "world"},
			},
			want: "hello? world!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringshelpers.CleanString(tt.args.initial, tt.args.stop); got != tt.want {
				t.Errorf("CleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}
