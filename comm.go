package steam

import (
	"bytes"
	"fmt"
	"math/rand"

	log "github.com/Sirupsen/logrus"
)

const (
	hInfoRequest                  = 'T'
	hInfoResponse                 = 'I'
	hPlayersInfoRequest           = 'U'
	hPlayersInfoChallengeResponse = 'A'
	hPlayersInfoResponse          = 'D'
)

type ServerType int

func (st *ServerType) unmarshalBinary(data []byte) error {
	switch data[0] {
	case 'd':
		*st = STDedicated
	case 'l':
		*st = STNonDedicated
	case 'p':
		*st = STProxy
	default:
		return errBadData
	}
	return nil
}

func (st ServerType) String() string {
	return serverTypeStrings[st]
}

const (
	STInvalid ServerType = iota
	STDedicated
	STNonDedicated
	STProxy
)

var serverTypeStrings = map[ServerType]string{
	STInvalid:      "Invalid",
	STDedicated:    "Dedicated",
	STNonDedicated: "Non Dedicated",
	STProxy:        "Proxy",
}

type Environment int

func (e *Environment) unmarshalBinary(data []byte) error {
	switch data[0] {
	case 'l':
		*e = ELinux
		return nil
	case 'w':
		*e = EWindows
		return nil
	case 'm', 'o':
		*e = EMac
		return nil
	default:
		return errBadData
	}
}

func (e Environment) String() string {
	return environmentStrings[e]
}

const (
	EInvalid Environment = iota
	ELinux
	EWindows
	EMac
)

var environmentStrings = map[Environment]string{
	EInvalid: "Invalid",
	ELinux:   "Linux",
	EWindows: "Windows",
	EMac:     "Mac",
}

type Visibility int

func (v *Visibility) unmarshalBinary(data []byte) error {
	switch data[0] {
	case 0:
		*v = VPublic
		return nil
	case 1:
		*v = VPrivate
		return nil
	default:
		return errBadData
	}
}

func (v Visibility) String() string {
	return visibilityStrings[v]
}

const (
	VInvalid Visibility = iota
	VPublic
	VPrivate
)

var visibilityStrings = map[Visibility]string{
	VInvalid: "Invalid",
	VPublic:  "Public",
	VPrivate: "Private",
}

type VAC int

func (v *VAC) unmarshalBinary(data []byte) error {
	switch data[0] {
	case 0:
		*v = VACUnsecured
		return nil
	case 1:
		*v = VACSecure
		return nil
	default:
		return errBadData
	}
}

func (v VAC) String() string {
	return vacStrings[v]
}

const (
	VACInvalid VAC = iota
	VACUnsecured
	VACSecure
)

var vacStrings = map[VAC]string{
	VACInvalid:   "Invalid",
	VACUnsecured: "Unsecured",
	VACSecure:    "Secured",
}

type infoRequest struct {
}

func (infoRequest) marshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	writeRequestPrefix(&buf)
	writeByte(&buf, hInfoRequest)
	writeString(&buf, "Source Engine Query")
	return buf.Bytes(), nil
}

type InfoResponse struct {
	Protocol    int
	Name        string
	Map         string
	Folder      string
	Game        string
	ID          int
	Players     int
	MaxPlayers  int
	Bots        int
	ServerType  ServerType
	Environment Environment
	Visibility  Visibility
	VAC         VAC
	Version     string

	Port    int
	SteamID int64

	SourceTVPort int
	SourceTVName string

	Keywords string
	GameID   int64
}

const (
	edfPort     = 0x80
	edfSteamID  = 0x10
	edfSourceTV = 0x40
	edfKeywords = 0x20
	edfGameID   = 0x01
)

func (r *InfoResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Print(err)
			err = e.(parseError)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hInfoResponse {
		triggerError(errBadData)
	}
	log.Debug("steam: info response header detected")
	r.Protocol = toInt(readByte(buf))
	r.Name = readString(buf)
	r.Map = readString(buf)
	r.Folder = readString(buf)
	r.Game = readString(buf)
	r.ID = toInt(readShort(buf))
	r.Players = toInt(readByte(buf))
	r.MaxPlayers = toInt(readByte(buf))
	r.Bots = toInt(readByte(buf))
	must(r.ServerType.unmarshalBinary(readBytes(buf, 1)))
	must(r.Environment.unmarshalBinary(readBytes(buf, 1)))
	must(r.Visibility.unmarshalBinary(readBytes(buf, 1)))
	must(r.VAC.unmarshalBinary(readBytes(buf, 1)))
	r.Version = readString(buf)
	// Check if EDF byte is present
	if buf.Len() < 1 {
		return nil
	}
	log.Debugf("steam: reading edf data (remaining bytes %v)", buf.Len())
	// EDF byte present
	edf := readByte(buf)
	if edf&edfPort != 0 {
		r.Port = toInt(readShort(buf))
	}
	if edf&edfSteamID != 0 {
		r.SteamID = readLongLong(buf)
	}
	if edf&edfSourceTV != 0 {
		r.SourceTVPort = toInt(readShort(buf))
		r.SourceTVName = readString(buf)
	}
	if edf&edfKeywords != 0 {
		r.Keywords = readString(buf)
	}
	if edf&edfGameID != 0 {
		r.GameID = readLongLong(buf)
		r.ID = int(r.GameID & 0xFFFFFF)
	}
	return nil
}

func (r *InfoResponse) String() string {
	return fmt.Sprintf("%v %v %v/%v (%v bots) %v", r.Name, r.Map, r.Players, r.MaxPlayers, r.Bots, r.VAC)
}

type playersInfoRequest struct {
	Challenge int
}

func (r playersInfoRequest) marshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	writeRequestPrefix(buf)
	writeByte(buf, hPlayersInfoRequest)
	writeLong(buf, int32(r.Challenge))
	return buf.Bytes(), nil
}

func isPlayersInfoChallengeResponse(b []byte) bool {
	return b[0] == hPlayersInfoChallengeResponse
}

type playersInfoChallengeResponse struct {
	Challenge int
}

func (r *playersInfoChallengeResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(parseError)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hPlayersInfoChallengeResponse {
		triggerError(errBadData)
	}
	log.Debug("steam: players info challenge response header detected")
	r.Challenge = toInt(readLong(buf))
	log.Debugf("steam: challenge number %#X", r.Challenge)
	return nil
}

type PlayersInfoResponse struct {
	Players []*Player
}

func (r *PlayersInfoResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(parseError)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hPlayersInfoResponse {
		triggerError(errBadData)
	}
	log.Debug("steam: players info response header detected")
	count := toInt(readByte(buf))
	log.Debugf("steam: received %v player info(s)", count)
	for i := 0; i < count; i++ {
		// Read the chunk index
		readByte(buf)
		var p Player
		p.Name = readString(buf)
		p.Score = toInt(readLong(buf))
		p.Duration = float64(readFloat(buf))
		r.Players = append(r.Players, &p)
	}
	return nil
}

type Player struct {
	Name     string
	Score    int
	Duration float64
}

type rconRequestType int32

const (
	rrtAuth      rconRequestType = 3
	rrtExecCmd   rconRequestType = 2
	rrtAuthResp  rconRequestType = 2
	rrtRespValue rconRequestType = 0
)

type rconRequest struct {
	size int32
	id   int32
	typ  rconRequestType
	body string
}

func (r *rconRequest) String() string {
	return fmt.Sprintf("%v %v %v %v", r.size, r.id, r.typ, r.body)
}

func newRCONRequest(typ rconRequestType, body string) *rconRequest {
	return &rconRequest{
		size: int32(len(body) + 10),
		id:   rand.Int31(),
		typ:  typ,
		body: body,
	}
}

func (r *rconRequest) marshalBinary() ([]byte, error) {
	log.Debugf("steam: rconRequest %v", r)
	var buf bytes.Buffer
	writeLong(&buf, r.size)
	writeLong(&buf, r.id)
	writeLong(&buf, int32(r.typ))
	buf.WriteString(r.body)
	writeNull(&buf)
	writeNull(&buf)
	return buf.Bytes(), nil
}

type rconResponse struct {
	size int32
	id   int32
	typ  rconRequestType
	body string
}

func (r *rconResponse) String() string {
	return fmt.Sprintf("%v %v %v %v", r.size, r.id, r.typ, r.body)
}

func (r *rconResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Print(err)
			err = e.(parseError)
		}
	}()
	buf := bytes.NewBuffer(data)
	r.size = readLong(buf)
	r.id = readLong(buf)
	r.typ = rconRequestType(readLong(buf))
	r.body = string(readBytes(buf, int(r.size-8)))
	log.Debugf("steam: rconResponse %v", r)
	return nil
}
