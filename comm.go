package steam

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const (
	hInfoRequest                  = 'T'
	hInfoResponse                 = 'I'
	hPlayersInfoRequest           = 'U'
	hPlayersInfoChallengeResponse = 'A'
	hPlayersInfoResponse          = 'D'
)

// ServerType indicates the type of the server.
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
	// STInvalid describes a invalid server type.
	STInvalid ServerType = iota

	// STDedicated indicates a dedicated server type.
	STDedicated
	// STNonDedicated indicates a non dedicated server type.
	STNonDedicated
	// STProxy indicates a proxy server type.
	STProxy
)

var serverTypeStrings = map[ServerType]string{
	STInvalid:      "Invalid",
	STDedicated:    "Dedicated",
	STNonDedicated: "Non Dedicated",
	STProxy:        "Proxy",
}

// Environment indicates the server's host environment.
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
	// EInvalid indicates a invalid host environment.
	EInvalid Environment = iota

	// ELinux indicates that the server is hosted on Linux.
	ELinux
	// EWindows indicates that the server is hosted on Windows.
	EWindows
	// EMac indicates that the server is hosted on Mac OS X.
	EMac
)

var environmentStrings = map[Environment]string{
	EInvalid: "Invalid",
	ELinux:   "Linux",
	EWindows: "Windows",
	EMac:     "Mac",
}

// Visibility indicates the visibility of the server.
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
	// VInvalid indicates a invalid visibility.
	VInvalid Visibility = iota

	// VPublic indicates a publicly visibly server.
	VPublic
	// VPrivate indicates a private server.
	VPrivate
)

var visibilityStrings = map[Visibility]string{
	VInvalid: "Invalid",
	VPublic:  "Public",
	VPrivate: "Private",
}

// VAC indicates the status of the Valve Anti Cheat system.
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
	// VACInvalid indicates a invalid VAC configuration.
	VACInvalid VAC = iota

	// VACUnsecured indicates a insecure server (VAC disabled).
	VACUnsecured
	// VACSecure indicates a secure server (VAC enabled).
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
	buf := new(bytes.Buffer)
	writeRequestPrefix(buf)
	writeByte(buf, hInfoRequest)
	writeString(buf, "Source Engine Query")
	return buf.Bytes(), nil
}

// InfoResponse represents a response to a info query.
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
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hInfoResponse {
		panic(errBadData)
	}
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
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hPlayersInfoChallengeResponse {
		panic(errBadData)
	}
	r.Challenge = toInt(readLong(buf))
	return nil
}

// PlayersInfoResponse represents a response to a player info query.
type PlayersInfoResponse struct {
	Players []*Player
}

func (r *PlayersInfoResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	buf := bytes.NewBuffer(data)
	header := readByte(buf)
	if header != hPlayersInfoResponse {
		panic(errBadData)
	}
	count := toInt(readByte(buf))
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

// Player represents a player entity in the server.
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
	buf := new(bytes.Buffer)
	writeLong(buf, r.size)
	writeLong(buf, r.id)
	writeLong(buf, int32(r.typ))
	buf.WriteString(r.body)
	writeNull(buf)
	writeNull(buf)
	return buf.Bytes(), nil
}

type rconResponse struct {
	size int32
	id   int32
	typ  rconRequestType
	body []byte
}

func (r *rconResponse) String() string {
	return fmt.Sprintf("%v %v %v %v", r.size, r.id, r.typ, string(r.body))
}

func (r *rconResponse) unmarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	buf := bytes.NewBuffer(data)
	r.size = readLong(buf)
	r.id = readLong(buf)
	r.typ = rconRequestType(readLong(buf))
	r.body = readBytes(buf, int(r.size-10))
	return nil
}

// TODO(anands): Temporary sample output.
//   CPU   NetIn   NetOut    Uptime  Maps   FPS   Players  Svms    +-ms   ~tick
//   10.0 241763.2 1518923.5   10419    58  127.98      16    3.72    1.56    0.36
// L 03/09/2018 - 08:59:17: rcon from "106.51.72.233:60415": command "stats"

// TODO(anands): Proper naming.
type StatsResponse struct {
	CPU         int
	NetIn       float64
	NetOut      float64
	Uptime      int
	Maps        int
	FPS         float64
	Players     int
	Svms        float64
	PlusMinusms float64
	Tick        float64
}

func (r *StatsResponse) unmarshalStatsRCONResponse(output string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	fields := strings.Fields(output)

	r.CPU = int(mustInterface(strconv.ParseFloat(fields[10], 64)).(float64))
	r.NetIn = mustInterface(strconv.ParseFloat(fields[11], 64)).(float64)
	r.NetOut = mustInterface(strconv.ParseFloat(fields[12], 64)).(float64)
	r.Uptime = mustInterface(strconv.Atoi(fields[13])).(int)
	r.Maps = mustInterface(strconv.Atoi(fields[14])).(int)
	r.FPS = mustInterface(strconv.ParseFloat(fields[15], 64)).(float64)
	r.Players = mustInterface(strconv.Atoi(fields[16])).(int)
	r.Svms = mustInterface(strconv.ParseFloat(fields[17], 64)).(float64)
	r.PlusMinusms = mustInterface(strconv.ParseFloat(fields[18], 64)).(float64)
	r.Tick = mustInterface(strconv.ParseFloat(fields[19], 64)).(float64)

	return nil
}
