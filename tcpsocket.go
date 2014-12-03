package steam

import (
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

type tcpSocket struct {
	conn net.Conn
}

func newTCPSocket(addr string) (*tcpSocket, error) {
	conn, err := Dial("tcp4", addr)
	if err != nil {
		glog.Errorf("steam: could not dial tcp to %v: %v", addr, err)
		return nil, err
	}
	glog.V(2).Infof("steam: created tcp connection to %v", addr)
	return &tcpSocket{conn}, nil
}

func (s *tcpSocket) close() {
	glog.V(2).Infof("steam: closing tcp connection to %v", s.conn.RemoteAddr())
	s.conn.Close()
}

func (s *tcpSocket) send(payload []byte) error {
	glog.V(1).Infof("steam: sending %v bytes tcp payload to %v", len(payload), s.conn.RemoteAddr())
	glog.V(2).Infof("steam: sending tcp payload to %v: %X", s.conn.RemoteAddr(), payload)
	if err := s.conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		glog.Errorf("steam: could not set write deadline for %v", s.conn.RemoteAddr())
		return err
	}
	n, err := s.conn.Write(payload)
	if err != nil {
		glog.Errorf("steam: error sending tcp data to %v: %v", s.conn.RemoteAddr(), err)
		return err
	}
	if n != len(payload) {
		return fmt.Errorf("steam: could not send full tcp request to %v", s.conn.RemoteAddr())
	}
	return nil
}

func (s *tcpSocket) receive() ([]byte, error) {
	glog.V(1).Infof("steam: trying to recieve bytes from %v", s.conn.RemoteAddr())
	if err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		glog.Errorf("steam: could not set read deadline for %v", s.conn.RemoteAddr())
		return nil, err
	}
	var buf [4096]byte
	n, err := s.conn.Read(buf[:])
	if err != nil {
		glog.Errorf("steam: could not read from %v: %v", s.conn.RemoteAddr(), err)
		return nil, err
	}
	glog.V(1).Infof("steam: received %v bytes from %v", n, s.conn.RemoteAddr())
	glog.V(2).Infof("steam: received tcp payload from %v: %X", s.conn.RemoteAddr(), buf[:n])
	return buf[:n], nil
}
