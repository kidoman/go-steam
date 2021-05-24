package steam

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"

	logrus "github.com/sirupsen/logrus"
)

type DialFn func(network, address string) (net.Conn, error)

// Server represents a Source engine game server.
type Server struct {
	addr string

	dial DialFn

	rconPassword string

	usock          *udpSocket
	udpInitialized bool

	rsock           *rconSocket
	rconInitialized bool

	mu sync.Mutex
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
func Connect(addr string, os ...*ConnectOptions) (_ *Server, err error) {
	s := &Server{
		addr: addr,
	}
	if len(os) > 0 {
		o := os[0]
		s.dial = o.Dial
		s.rconPassword = o.RCONPassword
	}
	if s.dial == nil {
		s.dial = (&net.Dialer{
			Timeout: 1 * time.Second,
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
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: could not open udp socket")
		return err
	}
	return nil
}

func (s *Server) initRCON() (err error) {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}
	log.WithFields(logrus.Fields{
		"addr": s.addr,
	}).Debug("steam: connecting rcon")
	if s.rsock, err = newRCONSocket(s.dial, s.addr); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: could not open tcp socket")
		return err
	}
	defer func() {
		if err != nil {
			s.rsock.close()
		}
	}()
	if err := s.authenticate(); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: could not authenticate")
		return err
	}
	s.rconInitialized = true
	return nil
}

func (s *Server) authenticate() error {
	log.WithFields(logrus.Fields{
		"addr": s.addr,
	}).Debug("steam: authenticating")
	req := newRCONRequest(rrtAuth, s.rconPassword)
	data, _ := req.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		return err
	}
	// Receive the empty response value
	data, err := s.rsock.receive()
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{
		"data": data,
	}).Debug("steam: received empty response")
	var resp rconResponse
	if err := resp.unmarshalBinary(data); err != nil {
		return err
	}
	if resp.typ != rrtRespValue || resp.id != req.id {
		return ErrInvalidResponseID
	}
	if resp.id != req.id {
		return ErrInvalidResponseType
	}
	// Receive the actual auth response
	data, err = s.rsock.receive()
	if err != nil {
		return err
	}
	if err := resp.unmarshalBinary(data); err != nil {
		return err
	}
	if resp.typ != rrtAuthResp || resp.id != req.id {
		return ErrRCONAuthFailed
	}
	log.Debug("steam: authenticated")
	return nil
}

// Close releases the resources associated with this server.
func (s *Server) Close() {
	if s.rconInitialized {
		s.rsock.close()
	}
	s.usock.close()
}

// Ping returns the RTT (round-trip time) to the server.
func (s *Server) Ping() (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	req, _ := infoRequest{}.marshalBinary()
	if err := s.usock.send(req); err != nil {
		return nil, err
	}
	log.Debug("receiving info response")
	data, err := s.usock.receive()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("could not receive info response")
		return nil, err
	}
	log.WithFields(logrus.Fields{
		"data": data,
	}).Debug("received info response")
	var res InfoResponse
	if err := res.unmarshalBinary(data); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("could not unmarshal info response")
		return nil, err
	}
	return &res, nil
}

// PlayersInfo retrieves player information from the server.
func (s *Server) PlayersInfo() (*PlayersInfoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.rconInitialized {
		return "", ErrRCONNotInitialized
	}
	req := newRCONRequest(rrtExecCmd, cmd)
	data, _ := req.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: sending rcon request")
		return "", err
	}
	// Send the mirror packet.
	reqMirror := newRCONRequest(rrtRespValue, "")
	data, _ = reqMirror.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: sending rcon mirror request")
		return "", err
	}
	var (
		buf       bytes.Buffer
		sawMirror bool
	)
	// Start receiving data.
	for {
		data, err := s.rsock.receive()
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Error("steam: receiving rcon response")
			return "", err
		}
		var resp rconResponse
		if err := resp.unmarshalBinary(data); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Error("steam: decoding response")
			return "", err
		}
		if resp.typ != rrtRespValue {
			return "", ErrInvalidResponseType
		}
		if !sawMirror && resp.id == reqMirror.id {
			sawMirror = true
			continue
		}
		if sawMirror {
			if bytes.Compare(resp.body, trailer) == 0 {
				break
			}
			return "", ErrInvalidResponseTrailer
		}
		if req.id != resp.id {
			return "", ErrInvalidResponseID
		}
		_, err = buf.Write(resp.body)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

var (
	trailer = []byte{0x00, 0x01, 0x00, 0x00}

	ErrRCONAuthFailed = errors.New("steam: authentication failed")

	ErrRCONNotInitialized     = errors.New("steam: rcon is not initialized")
	ErrInvalidResponseType    = errors.New("steam: invalid response type from server")
	ErrInvalidResponseID      = errors.New("steam: invalid response id from server")
	ErrInvalidResponseTrailer = errors.New("steam: invalid response trailer from server")
)

var log *logrus.Logger

func SetLog(l *logrus.Logger) {
	log = l
}

func init() {
	log = logrus.New()
	log.Out = ioutil.Discard
}
