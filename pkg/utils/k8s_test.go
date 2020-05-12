package utils

import (
	"testing"
)

func TestParseNameNS(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "parsesShouldWork",
			args:    args{input: "ns/name"},
			want:    "ns",
			want1:   "name",
			wantErr: false,
		},
		{
			name:    "parseShouldNotWork",
			args:    args{input: "notwork"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseNameNS(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNameNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseNameNS() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseNameNS() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
