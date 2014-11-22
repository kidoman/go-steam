package steam

import (
	"errors"
	"net"
	"time"

	"github.com/golang/glog"
)

type udpSocket struct {
	conn net.Conn
}

func newUDPSocket(addr string) (*udpSocket, error) {
	conn, err := net.DialTimeout("udp4", addr, time.Second)
	if err != nil {
		return nil, err
	}
	return &udpSocket{conn}, nil
}

func (s *udpSocket) close() {
	s.conn.Close()
}

func (s *udpSocket) send(payload []byte) error {
	glog.V(1).Infof("steam: sending %v bytes payload to %v", len(payload), s.conn.RemoteAddr())
	glog.V(2).Infof("steam: sending payload to %v: %X", s.conn.RemoteAddr(), payload)
	n, err := s.conn.Write(payload)
	if err != nil {
		return err
	}
	if n != len(payload) {
		return errors.New("steam: could not send full request to server")
	}
	return nil
}

func (s *udpSocket) receivePacket() ([]byte, error) {
	glog.V(1).Infof("steam: trying to recieve bytes from %v", s.conn.RemoteAddr())
	var buf [1500]byte
	s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := s.conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("steam: received %v bytes from %v", n, s.conn.RemoteAddr())
	glog.V(2).Infof("steam: received payload %v: %X", s.conn.RemoteAddr(), buf[:n])
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
