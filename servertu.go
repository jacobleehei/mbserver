package mbserver

import (
	"errors"
	"fmt"
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

		if (packet[0] != 0 && s.slaveID != 0) && packet[0] != s.slaveID {
			log.Printf("slave address mismatch: %v != %v\n", packet[0], s.slaveID)
			continue
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
	var (
		header  = make([]byte, 0)
		body    = make([]byte, 0)
		bodyLen *int
	)

	for {
		if bodyLen != nil && len(body) == *bodyLen {
			break
		}

		tmpBuf := make([]byte, 1)
		_, err := port.Read(tmpBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return header, err
		}

		if bodyLen != nil {
			body = append(body, tmpBuf...)
		} else {
			header = append(header, tmpBuf...)
			if len(header) == 6 {
				remindPacketLength, err := calRtuPacketBodyLength(header)
				if err != nil {
					return nil, err
				}
				bodyLen = &remindPacketLength
			}
		}
	}

	return append(header, body...), nil
}

type FuncCode int

const (
	FuncCodeReadCoils              FuncCode = 1
	FuncCodeReadDiscreteInputs     FuncCode = 2
	FuncCodeReadHoldingRegisters   FuncCode = 3
	FuncCodeReadInputRegisters     FuncCode = 4
	FuncCodeWriteSingleCoil        FuncCode = 5
	FuncCodeWriteSingleRegister    FuncCode = 6
	FuncCodeWriteMultipleCoils     FuncCode = 15
	FuncCodeWriteMultipleRegisters FuncCode = 16
)

// modbus rtu header format:
//
//	byte 0: slave address
//	byte 1: function code
//	byte 2, 3: starting address
//	byte 4, 5:  quantity of coils/registers
func calRtuPacketBodyLength(headerBytes []byte) (int, error) {
	if len(headerBytes) < 6 {
		return 0, errors.New("header len < 6")
	}

	funcCode := FuncCode(headerBytes[1])
	switch funcCode {
	case FuncCodeReadCoils, FuncCodeReadDiscreteInputs, FuncCodeReadHoldingRegisters, FuncCodeReadInputRegisters:
		return 2, nil // checksum
	case FuncCodeWriteSingleCoil, FuncCodeWriteSingleRegister:
		return 2, nil // checksum
	case FuncCodeWriteMultipleCoils, FuncCodeWriteMultipleRegisters:
		return (int(headerBytes[4])*256+int(headerBytes[5]))*2 + 2, nil
	}

	return 0, fmt.Errorf("unknown function code: %v", funcCode)
}
