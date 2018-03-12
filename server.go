package steam

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"

	logrus "github.com/Sirupsen/logrus"
)

// The DialFn type is an adapter to allow the use of
// a custom network dialing mechanism when required.
// For example, this will come useful inside a environment
// like AppEngine which does not permit direct socket
// connections and requires the usage of a custom dialer.
type DialFn func(network, address string) (net.Conn, error)

type connectOptions struct {
	dialFn       DialFn
	rconPassword string
}

// ConnectOption configures how we set up the connection.
type ConnectOption func(*connectOptions)

// WithDialFn returns a ConnectOption which sets a dialFn for establishing
// connection to the server.
func WithDialFn(fn DialFn) ConnectOption {
	return func(o *connectOptions) {
		o.dialFn = fn
	}
}

// WithRCONPassword returns a ConnectOption which sets a rcon password for
// authenticating the connection to the server.
func WithRCONPassword(password string) ConnectOption {
	return func(o *connectOptions) {
		o.rconPassword = password
	}
}

// Server represents a Source engine game server.
type Server struct {
	addr string

	opts connectOptions

	usock          *udpSocket
	udpInitialized bool

	rsock           *rconSocket
	rconInitialized bool

	mu sync.Mutex
}

// Connect to the source server.
func Connect(addr string, opts ...ConnectOption) (_ *Server, err error) {
	s := Server{
		addr: addr,
	}

	for _, opt := range opts {
		opt(&s.opts)
	}

	if s.opts.dialFn == nil {
		s.opts.dialFn = (&net.Dialer{
			Timeout: 1 * time.Second,
		}).Dial
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			s.usock.close()
		}
	}()

	if s.opts.rconPassword == "" {
		return &s, nil
	}

	if err := s.initRCON(); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) String() string {
	return s.addr
}

func (s *Server) init() error {
	if s.addr == "" {
		return errors.New("steam: server needs a address")
	}

	var err error
	if s.usock, err = newUDPSocket(s.opts.dialFn, s.addr); err != nil {
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

	if s.rsock, err = newRCONSocket(s.opts.dialFn, s.addr); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			s.rsock.close()
		}
	}()

	if err := s.authenticate(); err != nil {
		return err
	}

	s.rconInitialized = true

	return nil
}

func (s *Server) authenticate() error {
	log.WithFields(logrus.Fields{
		"addr": s.addr,
	}).Debug("steam: authenticating")

	req := newRCONRequest(rrtAuth, s.opts.rconPassword)
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

	start := time.Now()

	req, _ := infoRequest{}.marshalBinary()
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
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"data": data,
	}).Debug("received info response")

	var res InfoResponse
	if err := res.unmarshalBinary(data); err != nil {
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

	log.WithFields(logrus.Fields{
		"addr": s.addr,
		"id":   req.id,
	}).Debug("steam: sending rcon request")

	if err := s.rsock.send(data); err != nil {
		return "", err
	}

	log.WithFields(logrus.Fields{
		"addr": s.addr,
		"id":   req.id,
	}).Debug("steam: sent rcon request")

	// Send the mirror packet.
	reqMirror := newRCONRequest(rrtRespValue, "")
	data, _ = reqMirror.marshalBinary()

	log.WithFields(logrus.Fields{
		"addr": s.addr,
		"id":   reqMirror.id,
	}).Debug("steam: sending rcon mirror request")

	if err := s.rsock.send(data); err != nil {
		return "", err
	}

	log.WithFields(logrus.Fields{
		"addr": s.addr,
		"id":   reqMirror.id,
	}).Debug("steam: sent rcon mirror request")

	var (
		buf       bytes.Buffer
		sawMirror bool
	)

	// Start receiving data.
	for {
		data, err := s.rsock.receive()
		if err != nil {
			return "", err
		}

		log.WithFields(logrus.Fields{
			"addr": s.addr,
		}).Debug("steam: received rcon response")

		var resp rconResponse
		if err := resp.unmarshalBinary(data); err != nil {
			return "", err
		}

		if resp.typ != rrtRespValue {
			return "", ErrInvalidResponseType
		}

		if !sawMirror && resp.id == reqMirror.id {
			sawMirror = true

			log.WithFields(logrus.Fields{
				"addr": s.addr,
				"id":   resp.id,
			}).Debug("steam: received mirror request")

			continue
		}
		if sawMirror {
			if bytes.Compare(resp.body, trailer) == 0 {
				log.WithFields(logrus.Fields{
					"addr": s.addr,
				}).Debug("steam: received mirror trailer")

				break
			}
			return "", ErrInvalidResponseTrailer
		}

		if resp.id != req.id {
			return "", ErrInvalidResponseID
		}

		if _, err := buf.Write(resp.body); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

var trailer = []byte{0x00, 0x01, 0x00, 0x00}

// Errors introduced by the steam client.
var (
	ErrRCONAuthFailed = errors.New("steam: authentication failed")

	ErrRCONNotInitialized     = errors.New("steam: rcon is not initialized")
	ErrInvalidResponseType    = errors.New("steam: invalid response type from server")
	ErrInvalidResponseID      = errors.New("steam: invalid response id from server")
	ErrInvalidResponseTrailer = errors.New("steam: invalid response trailer from server")
)

var log *logrus.Logger

// SetLog overrides the logger used by the steam client.
func SetLog(l *logrus.Logger) {
	log = l
}

func init() {
	log = logrus.New()
	log.Out = ioutil.Discard
}

// Stats retrieves server stats.
func (s *Server) Stats() (*StatsResponse, error) {
	log.Debug("receiving stats response")

	output, err := s.Send("stats")
	if err != nil {
		return nil, err
	}

	log.Debug("unmarshaling stats response")

	var res StatsResponse
	if err := res.unmarshalStatsRCONResponse(output); err != nil {
		return nil, err
	}

	return &res, nil
}
