package steam

import (
	"errors"
	"net"
	"time"

	"github.com/golang/glog"
)

type DialFn func(network, address string) (net.Conn, error)

// Server represents a Source engine game server.
type Server struct {
	addr string

	dial DialFn

	rconPassword string

	usock          *udpSocket
	udpInitialized bool

	tsock           *tcpSocket
	rconInitialized bool
}

// ConnectOptions describes the various connections options.
type ConnectOptions struct {
	// Default will use net.Dialer.Dial. You can override the same by
	// providing your own.
	Dial DialFn

	// RCON password.
	RCONPassword string
}

// Connect to the source server.
func Connect(addr string, os ...*ConnectOptions) (s *Server, err error) {
	s = &Server{
		addr: addr,
	}
	if len(os) > 0 {
		o := os[0]
		s.dial = o.Dial
		s.rconPassword = o.RCONPassword
	}
	if s.dial == nil {
		s.dial = (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	if s.rconPassword == "" {
		return s, nil
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

func (s *Server) String() string {
	return s.addr
}

func (s *Server) init() error {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}
	var err error
	if s.usock, err = newUDPSocket(s.dial, s.addr); err != nil {
		glog.Errorf("server: could not create udp socket to %v: %v", s.addr, err)
		return err
	}
	return nil
}

func (s *Server) initRCON() (err error) {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}
	if s.tsock, err = newTCPSocket(s.dial, s.addr); err != nil {
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

func (s *Server) authenticate() error {
	req := newRCONRequest(rrtAuth, s.rconPassword)
	glog.V(2).Infof("steam: sending rcon auth request: %v", req)
	data, _ := req.marshalBinary()
	if err := s.tsock.send(data); err != nil {
		return err
	}
	// Receive the empty response value
	data, err := s.tsock.receive()
	if err != nil {
		return err
	}
	var resp rconResponse
	if err := resp.unmarshalBinary(data); err != nil {
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
	if err := resp.unmarshalBinary(data); err != nil {
		return err
	}
	glog.V(2).Infof("steam: received response %v", resp)
	if resp.typ != rrtAuthResp || resp.id != req.id {
		return ErrRCONAuthFailed
	}
	return nil
}

// Close releases the resources associated with this server.
func (s *Server) Close() {
	if s.rconInitialized {
		s.tsock.close()
	}
	s.usock.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *Server) Ping() (time.Duration, error) {
	req, _ := infoRequest{}.marshalBinary()
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
	req, _ := infoRequest{}.marshalBinary()
	if err := s.usock.send(req); err != nil {
		return nil, err
	}
	data, err := s.usock.receive()
	if err != nil {
		return nil, err
	}
	var res InfoResponse
	if err := res.unmarshalBinary(data); err != nil {
		return nil, err
	}
	return &res, nil
}

// PlayersInfo retrieves player information from the server.
func (s *Server) PlayersInfo() (*PlayersInfoResponse, error) {
	// Send the challenge request
	req, _ := playersInfoRequest{}.marshalBinary()
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
		if err := challangeRes.unmarshalBinary(data); err != nil {
			return nil, err
		}
		// Send a new request with the proper challenge number
		req, _ = playersInfoRequest{challangeRes.Challenge}.marshalBinary()
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
	if err := res.unmarshalBinary(data); err != nil {
		return nil, err
	}
	return &res, nil
}

// Send RCON command to the server.
func (s *Server) Send(cmd string) (string, error) {
	if !s.rconInitialized {
		return "", ErrRCONNotInitialized
	}
	req := newRCONRequest(rrtExecCmd, cmd)
	glog.V(2).Infof("steam: sending rcon exec command request: %v", req)
	data, _ := req.marshalBinary()
	if err := s.tsock.send(data); err != nil {
		return "", err
	}
	data, err := s.tsock.receive()
	if err != nil {
		return "", err
	}
	var resp rconResponse
	if err := resp.unmarshalBinary(data); err != nil {
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
