package steam

import (
	"errors"
	"time"
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

	if err := s.socket.send(data); err != nil {
		return nil, err
	}
	b, err := s.socket.receive()
	if err != nil {
		return nil, err
	}

	res := new(InfoResponse)
	if err := res.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return res, nil
}
