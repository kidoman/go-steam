package steam

import (
	"errors"
	"time"

	"github.com/golang/glog"
)

// Server represents a Source server.
type server struct {
	addr         string
	rconPassword string

	usock *udpSocket
	tsock *tcpSocket

	rconInitialized bool
}

func Connect(addr string) (*server, error) {
	s := &server{
		addr: addr,
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func ConnectAuth(addr, rconPassword string) (s *server, err error) {
	s = &server{
		addr:         addr,
		rconPassword: rconPassword,
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			s.usock.close()
		}
	}()
	if err := s.initRCON(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *server) String() string {
	return s.addr
}

func (s *server) init() error {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}
	var err error
	if s.usock, err = newUDPSocket(s.addr); err != nil {
		glog.Errorf("server: could not create udp socket to %v: %v", s.addr, err)
		return err
	}
	return nil
}

func (s *server) initRCON() (err error) {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}
	if s.tsock, err = newTCPSocket(s.addr); err != nil {
		glog.Errorf("server: could not create tcp socket to %v: %v", s.addr, err)
		return err
	}
	defer func() {
		if err != nil {
			s.tsock.close()
		}
	}()
	if err := s.authenticate(); err != nil {
		glog.Errorf("server: could not authenticate rcon to %v: %v", s.addr, err)
		return err
	}
	s.rconInitialized = true
	return nil
}

func (s *server) authenticate() error {
	req := newRCONRequest(rrtAuth, s.rconPassword)
	glog.V(2).Infof("steam: sending rcon auth request: %v", req)
	data, _ := req.MarshalBinary()
	if err := s.tsock.send(data); err != nil {
		return err
	}
	// Receive the empty response value
	data, err := s.tsock.receive()
	if err != nil {
		return err
	}
	var resp rconResponse
	if err := resp.UnmarshalBinary(data); err != nil {
		return err
	}
	glog.V(2).Infof("steam: received response %v", resp)
	if resp.typ != rrtRespValue || resp.id != req.id {
		return ErrInvalidResponseID
	}
	if resp.id != req.id {
		return ErrInvalidResponseType
	}
	// Receive the actual auth response
	data, err = s.tsock.receive()
	if err != nil {
		return err
	}
	if err := resp.UnmarshalBinary(data); err != nil {
		return err
	}
	glog.V(2).Infof("steam: received response %v", resp)
	if resp.typ != rrtAuthResp || resp.id != req.id {
		return ErrRCONAuthFailed
	}
	return nil
}

// Close releases the resources associated with this server.
func (s *server) Close() {
	if s.rconInitialized {
		s.tsock.close()
	}
	s.usock.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *server) Ping() (time.Duration, error) {
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
func (s *server) Info() (*InfoResponse, error) {
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
func (s *server) PlayersInfo() (*PlayersInfoResponse, error) {
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

func (s *server) Send(cmd string) (string, error) {
	if !s.rconInitialized {
		return "", ErrRCONNotInitialized
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
	glog.V(2).Infof("steam: received response %v", resp)
	if resp.typ != rrtRespValue {
		return "", ErrInvalidResponseType
	}
	if req.id != resp.id {
		return "", ErrInvalidResponseID
	}
	return resp.body, nil
}

var (
	ErrRCONAuthFailed = errors.New("steam: authentication failed")

	ErrRCONNotInitialized  = errors.New("steam: rcon is not initialized")
	ErrInvalidResponseType = errors.New("steam: invalid response from server")
	ErrInvalidResponseID   = errors.New("steam: invalid response from server")
)
