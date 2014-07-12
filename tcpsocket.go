package steam

import (
	"net"

	"github.com/golang/glog"
)

type tcpSocket struct {
	conn  *net.TCPConn
	raddr *net.TCPAddr
}

func newTcpSocket(addr string) (*tcpSocket, error) {

	raddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		glog.Errorf("steam: could not resolve tcp addr coz: %v", err.Error())
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, raddr)
	if err != nil {
		glog.Errorf("steam: could not dial tcp coz: %v", err.Error())
		return nil, err
	}

	glog.V(2).Infof("steam: succesfully created tcp connection. conn:%v, raddr: %v", conn, raddr)
	return &tcpSocket{conn, raddr}, nil
}

func (s *tcpSocket) close() {
	glog.V(2).Infof("steam: closing tcp connection")
	s.conn.Close()
}

func (s *tcpSocket) send(payload []byte) error {
	glog.V(2).Infof("steam: sending payload %v", payload)
	_, err := s.conn.Write(payload)
	if err != nil {
		glog.V(2).Infof("steam: error sending data: %v", err.Error())
	}
	return nil
}

func (s *tcpSocket) receive() ([]byte, error) {
	var buf [4095]byte
	glog.V(1).Infof("steam: reading from %v", s.raddr)

	n, err := s.conn.Read(buf[:])
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("steam: received %v bytes from %v", n, s.raddr)

	return buf[:n], nil
}
