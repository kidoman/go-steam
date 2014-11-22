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

	RCONPassword string

	usock *udpSocket
	tsock *tcpSocket

	initialized     bool
	rconInitialized bool
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
		glog.Errorf("server: could not create udp socket to %v: %v", s.Addr, err)
		return err
	}
	if s.RCONPassword == "" {
		if s.tsock, err = newTCPSocket(s.Addr); err != nil {
			glog.Errorf("server: could not create tcp socket to %v: %v", s.Addr, err)
			return err
		}
		ok, err := s.authenticate()
		if err != nil {
			glog.Errorf("server: could not authenticate rcon to %v: %v", s.Addr, err)
			return err
		}
		if !ok {
			return errors.New("steam: rcon authentication failed")
		}
		s.rconInitialized = true
	}
	s.initialized = true
	return nil
}

func (s *Server) authenticate() (bool, error) {
	req := newRCONRequest(rrtAuth, s.RCONPassword)
	glog.V(2).Infof("steam: sending rcon auth request: %v", req)
	data, _ := req.MarshalBinary()
	if err := s.tsock.send(data); err != nil {
		return false, err
	}
	data, err := s.tsock.receive()
	if err != nil {
		return false, err
	}
	var resp rconResponse
	if err := resp.UnmarshalBinary(data); err != nil {
		return false, err
	}
	if resp.id != req.id {
		return false, nil
	}
	return true, nil
}

// Close releases the resources associated with this server.
func (s *Server) Close() {
	if !s.initialized {
		return
	}

	s.usock.close()
	s.tsock.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *Server) Ping() (time.Duration, error) {
	if err := s.init(); err != nil {
		return 0, err
	}
	req, _ := infoRequest{}.MarshalBinary()
	start := time.Now()
	s.usock.send(req)
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
	req, _ := infoRequest{}.MarshalBinary()
	if err := s.usock.send(req); err != nil {
		return nil, err
	}
	data, err := s.usock.receive()
	if err != nil {
		return nil, err
	}
	var res InfoResponse
	if err := res.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &res, nil
}

// PlayersInfo retrieves player information from the server.
func (s *Server) PlayersInfo() (*PlayersInfoResponse, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	// Send the challenge request
	req, _ := playersInfoRequest{}.MarshalBinary()
	if err := s.usock.send(req); err != nil {
		return nil, err
	}
	data, err := s.usock.receive()
	if err != nil {
		return nil, err
	}
	if isPlayersInfoChallengeResponse(data) {
		// Parse the challenge response
		var challangeRes playersInfoChallengeResponse
		if err := challangeRes.UnmarshalBinary(data); err != nil {
			return nil, err
		}
		// Send a new request with the proper challenge number
		req, _ = playersInfoRequest{challangeRes.Challenge}.MarshalBinary()
		if err := s.usock.send(req); err != nil {
			return nil, err
		}
		data, err = s.usock.receive()
		if err != nil {
			return nil, err
		}
	}
	// Parse the return value
	var res PlayersInfoResponse
	if err := res.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Server) Send(cmd string) (string, error) {
	if err := s.init(); err != nil {
		return "", err
	}
	if !s.rconInitialized {
		return "", errors.New("steam: rcon is not initialized")
	}
	req := newRCONRequest(rrtExecCmd, cmd)
	glog.V(2).Infof("steam: sending rcon exec command request: %v", req)
	data, _ := req.MarshalBinary()
	if err := s.tsock.send(data); err != nil {
		return "", err
	}
	data, err := s.tsock.receive()
	if err != nil {
		return "", err
	}
	var resp rconResponse
	if err := resp.UnmarshalBinary(data); err != nil {
		return "", err
	}
	if req.id != resp.id {
		err := errors.New("steam: response id does not match request id")
		return "", err
	}
	return resp.body, nil
}
