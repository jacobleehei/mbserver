package mbserver

import (
	"testing"

	"github.com/goburrow/serial"
)

func TestReadFullRTUPacket(t *testing.T) {
	t.Logf("start test")
	port, err := serial.Open(&serial.Config{Address: "/dev/tty4"})
	if err != nil {
		t.Fatal(err)
	}
	defer port.Close()

	b, err := readFullRTUPacket(port)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("b = %v", b)
}

func TestCalRtuPacketBodyLength(t *testing.T) {
	type args struct {
		headerBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "read coils",
			args: args{
				headerBytes: []byte{0x01, 0x01, 0x01, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "read discrete inputs",
			args: args{
				headerBytes: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "read holding registers",
			args: args{
				headerBytes: []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "read input registers",
			args: args{
				headerBytes: []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "write single coil",
			args: args{
				headerBytes: []byte{0x01, 0x05, 0x00, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "write single register",
			args: args{
				headerBytes: []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x00},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "write multiple coils",
			args: args{
				headerBytes: []byte{0x01, 0x0f, 0x00, 0x00, 0x00, 0x10},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "write multiple registers",
			args: args{
				headerBytes: []byte{0x01, 0x10, 0x00, 0x00, 0x01, 0x0},
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calRtuPacketBodyLength(tt.args.headerBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("calRtuPacketBodyLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calRtuPacketBodyLength() = %v, want %v", got, tt.want)
			}
			t.Logf("calRtuPacketBodyLength() = %v", got)
		})
	}
}
