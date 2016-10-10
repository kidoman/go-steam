package steam

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type udpSocket struct {
	conn net.Conn
}

func newUDPSocket(dial DialFn, addr string) (*udpSocket, error) {
	conn, err := dial("udp", addr)
	if err != nil {
		return nil, err
	}

	return &udpSocket{conn}, nil
}

func (s *udpSocket) close() {
	s.conn.Close()
}

func (s *udpSocket) send(payload []byte) error {
	n, err := s.conn.Write(payload)
	if err != nil {
		return err
	}
	if n != len(payload) {
		return fmt.Errorf("steam: could not send full udp request to %v", s.conn.RemoteAddr())
	}
	return nil
}

func (s *udpSocket) receivePacket() ([]byte, error) {
	if err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return nil, err
	}
	buf := make([]byte, 1500)
	n, err := s.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *udpSocket) receive() ([]byte, error) {
	buf, err := s.receivePacket()
	if err != nil {
		return nil, err
	}
	if buf[0] == 0xFE {
		return nil, errors.New("steam: cannot handle split packets")
	}
	return buf[4:], nil
}
