package steam

import (
	"errors"
	"time"

	"github.com/golang/glog"
)

// Server represents a Source server.
type Server struct {
	// IP:Port combination designating a single server.
	Addr string

	socket *socket

	initialized bool
}

func (s *Server) init() error {
	if s.initialized {
		return nil
	}

	if s.Addr == "" {
		return errors.New("steam: server needs a address")
	}

	var err error
	if s.socket, err = newSocket(s.Addr); err != nil {
		return err
	}

	s.initialized = true
	return nil
}

// Close releases the resources associated with this server.
func (s *Server) Close() {
	if !s.initialized {
		return
	}

	s.socket.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *Server) Ping() (time.Duration, error) {
	if err := s.init(); err != nil {
		return 0, err
	}

	data, err := InfoRequest{}.MarshalBinary()
	if err != nil {
		return 0, err
	}

	glog.V(3).Infof("steam: sending data %v via socket in ping", data)
	start := time.Now()
	s.socket.send(data)
	if _, err := s.socket.receive(); err != nil {
		return 0, err
	}

	elapsed := time.Since(start)
	return elapsed, nil
}

// Info retrieves server information.
func (s *Server) Info() (*InfoResponse, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	data, err := InfoRequest{}.MarshalBinary()
	if err != nil {
		return nil, err
	}

	b, err := s.sendAndRecieve(data)
	if err != nil {
		return nil, err
	}

	res := new(InfoResponse)
	if err := res.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return res, nil
}

// PlayerInfo retrieves players information on the server.
func (s *Server) PLayerInfo() (*A2SPlayersResponse, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	data := ChallengeRequest{}.MarshalBinary()

	b, err := s.sendAndRecieve(data)
	if err != nil {
		return nil, err
	}

	challengeRes := ChallengeResponse(b)
	data = A2SPlayerRequest{}.MarshalBinaryFromChallenge(challengeRes)

	b, err = s.sendAndRecieve(data)
	if err != nil {
		return nil, err
	}

	a2sPlayerRes := new(A2SPlayersResponse)

	if err := a2sPlayerRes.UnMarshalBinary(b); err != nil {
		return nil, err
	}

	return a2sPlayerRes, nil
}

func (s *Server) sendAndRecieve(data []byte) ([]byte, error) {
	glog.V(3).Infof("steam: sending data %v via socket in info", data)
	if err := s.socket.send(data); err != nil {
		return nil, err
	}

	b, err := s.socket.receive()
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("steam: received data %v via socket", b)
	return b, nil
}
