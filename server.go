package steam

import (
	"errors"
	"time"
)

// Server represents a Source server.
type Server struct {
	// IP:Port combination designating a single server.
	Addr string

	usock *udpSocket

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
	if s.usock, err = newUDPSocket(s.Addr); err != nil {
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

	s.usock.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *Server) Ping() (time.Duration, error) {
	if err := s.init(); err != nil {
		return 0, err
	}
	data, err := infoRequest{}.MarshalBinary()
	if err != nil {
		return 0, err
	}
	start := time.Now()
	s.usock.send(data)
	if _, err := s.usock.receive(); err != nil {
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

	data, err := infoRequest{}.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := s.usock.send(data); err != nil {
		return nil, err
	}
	b, err := s.usock.receive()
	if err != nil {
		return nil, err
	}

	res := new(InfoResponse)
	if err := res.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return res, nil
}

// PlayersInfo retrieves player information from the server.
func (s *Server) PlayersInfo() (*PlayersInfoResponse, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	// Send the challenge request
	data, err := playersInfoRequest{}.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := s.usock.send(data); err != nil {
		return nil, err
	}
	b, err := s.usock.receive()
	if err != nil {
		return nil, err
	}

	if isPlayersInfoChallengeResponse(b) {
		// Parse the challenge response
		challangeRes := new(playersInfoChallengeResponse)
		if err := challangeRes.UnmarshalBinary(b); err != nil {
			return nil, err
		}

		// Send a new request with the proper challenge number
		data, err = playersInfoRequest{challangeRes.Challenge}.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if err := s.usock.send(data); err != nil {
			return nil, err
		}
		b, err = s.usock.receive()
		if err != nil {
			return nil, err
		}
	}

	// Parse the return value
	res := new(PlayersInfoResponse)
	if err := res.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return res, nil
}
