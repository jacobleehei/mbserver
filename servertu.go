package mbserver

import (
	"io"
	"log"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)

	s.portsWG.Add(1)
	go func() {
		defer s.portsWG.Done()
		s.acceptSerialRequests(port)
	}()

	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
SkipFrameError:
	for {
		select {
		case <-s.portsCloseChan:
			return
		default:
		}

		packet, err := readFullRTUPacket(port)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}

		frame, err := NewRTUFrame(packet)
		if err != nil {
			log.Printf("bad serial frame error %v\n", err)
			//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous
			//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
			log.Printf("Keep the RTU server running!!\n")
			continue SkipFrameError
			//return
		}

		request := &Request{port, frame}

		s.requestChan <- request
	}
}

func readFullRTUPacket(port serial.Port) ([]byte, error) {
	packet := make([]byte, 0)
	for {
		if len(packet) >= 3 {
			packetLength := packet[2] + 5 // including header, slave address, function code and CRC
			if len(packet) >= int(packetLength) {
				return packet[:packetLength], nil
			}
		}
		tmpBuf := make([]byte, 1)
		_, err := port.Read(tmpBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return packet, err
		}

		packet = append(packet, tmpBuf...)
	}
}
