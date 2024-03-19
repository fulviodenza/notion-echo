package parser

import "testing"

func TestParseSchedule(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "return parsed schedule",
			args: args{
				s: "15:30",
			},
			want:    "CRON_TZ=Europe/Rome 30 15 * * *",
			wantErr: false,
		},
		{
			name: "return parsed schedule",
			args: args{
				s: "50:30",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "return parsed schedule",
			args: args{
				s: "24:30",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateSchedule(tt.args.s, "Europe/Rome")
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}
