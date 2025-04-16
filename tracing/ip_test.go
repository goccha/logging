package tracing

import (
	"testing"
)

func TestIpHeaders_Prepend(t *testing.T) {
	type args struct {
		header IpHeader
	}
	tests := []struct {
		name string
		h    IpHeaders
		args args
		want IpHeaders
	}{
		{
			name: "prepend",
			h:    DefaultIpHeaders(),
			args: args{
				header: XRealIp(),
			},
			want: []IpHeader{
				XRealIp(),
				Forwarded(),
				XForwardedFor(),
				RemoteAddr(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.Prepend(tt.args.header); len(got) != len(tt.want) {
				t.Errorf("Prepend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIpHeaders_Append(t *testing.T) {
	type args struct {
		header IpHeader
	}
	tests := []struct {
		name string
		h    IpHeaders
		args args
		want IpHeaders
	}{
		{
			name: "append",
			h:    DefaultIpHeaders(),
			args: args{
				header: XRealIp(),
			},
			want: []IpHeader{
				Forwarded(),
				XForwardedFor(),
				RemoteAddr(),
				XRealIp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.Append(tt.args.header); len(got) != len(tt.want) {
				t.Errorf("Append() = %v, want %v", got, tt.want)
			}
		})
	}
}
