package steam

import (
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

type udpSocket struct {
	conn net.Conn
}

func newUDPSocket(dial DialFn, addr string) (*udpSocket, error) {
	conn, err := dial("udp", addr)
	if err != nil {
		log.Errorf("steam: could not dial udp to %v: %v", addr, err)
		return nil, err
	}
	log.Debugf("steam: created udp connection to %v", addr)
	return &udpSocket{conn}, nil
}

func (s *udpSocket) close() {
	log.Debugf("steam: closing udp connection to %v", s.conn.RemoteAddr())
	s.conn.Close()
}

func (s *udpSocket) send(payload []byte) error {
	log.Debugf("steam: sending %v bytes udp payload to %v", len(payload), s.conn.RemoteAddr())
	log.Debugf("steam: sending udp payload to %v: %X", s.conn.RemoteAddr(), payload)
	n, err := s.conn.Write(payload)
	if err != nil {
		log.Errorf("steam: error sending udp data to %v: %v", s.conn.RemoteAddr(), err)
		return err
	}
	if n != len(payload) {
		return fmt.Errorf("steam: could not send full udp request to %v", s.conn.RemoteAddr())
	}
	return nil
}

func (s *udpSocket) receivePacket() ([]byte, error) {
	log.Debugf("steam: trying to recieve bytes from %v", s.conn.RemoteAddr())
	if err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		log.Errorf("steam: could not set read deadline for %v", s.conn.RemoteAddr())
		return nil, err
	}
	var buf [1500]byte
	n, err := s.conn.Read(buf[:])
	if err != nil {
		log.Errorf("steam: could not read from %v: %v", s.conn.RemoteAddr(), err)
		return nil, err
	}
	log.Debugf("steam: received %v bytes from %v", n, s.conn.RemoteAddr())
	log.Debugf("steam: received udp payload from %v: %X", s.conn.RemoteAddr(), buf[:n])
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
