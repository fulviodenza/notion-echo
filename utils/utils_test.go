package utils

import "testing"

func TestValidateTime(t *testing.T) {
	type args struct {
		times []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Correctly validate times",
			args: args{
				times: []string{"14", "30"},
			},
			want: true,
		},
		{
			name: "Hours too high",
			args: args{
				times: []string{"24", "30"},
			},
			want: false,
		},
		{
			name: "Hours too low",
			args: args{
				times: []string{"-1", "30"},
			},
			want: false,
		},
		{
			name: "Minutes too high",
			args: args{
				times: []string{"15", "61"},
			},
			want: false,
		},
		{
			name: "Minutes too low",
			args: args{
				times: []string{"15", "-1"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateTime(tt.args.times, "Europe/Rome"); got != tt.want {
				t.Errorf("ValidateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
