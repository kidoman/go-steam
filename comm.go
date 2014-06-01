package steam

import (
	"bytes"
	"fmt"

	"github.com/golang/glog"
)

type Header byte

const (
	HInfo Header = 0x49
)

type ServerType int

func (st *ServerType) UnmarshalBinary(data []byte) error {
	glog.V(3).Infof("unmarshalling binary %v for server type ", data)
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

func (e *Environment) UnmarshalBinary(data []byte) error {
	glog.V(3).Infof("unmarshalling binary %v for env ", data)
	switch data[0] {
	case 'l':
		*e = ELinux
	case 'w':
		*e = EWindows
	case 'm', 'o':
		*e = EMac
	default:
		return errBadData
	}

	return nil
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

func (v *Visibility) UnmarshalBinary(data []byte) error {
	glog.V(3).Infof("unmarshalling binary %v for visibility ", data)
	switch data[0] {
	case 0:
		*v = VPublic
	case 1:
		*v = VPrivate
	default:
		return errBadData
	}

	return nil
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

func (v *VAC) UnmarshalBinary(data []byte) error {
	switch data[0] {
	case 0:
		*v = VACUnsecured
	case 1:
		*v = VACSecure
	default:
		return errBadData
	}

	return nil
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

type InfoRequest struct {
}

func (InfoRequest) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	writeRequestPrefix(buf)
	writeByte(buf, 'T')
	writeString(buf, "Source Engine Query")

	glog.V(3).Infof("marshaled binary. buffer: %v", buf)
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

func (r *InfoResponse) UnmarshalBinary(data []byte) (err error) {
	glog.V(3).Infof("unmarshalling binary %v", data)
	defer func() {
		if e := recover(); e != nil {
			err = e.(parseError)
		}
	}()

	buf := bytes.NewBuffer(data)

	r.Protocol = toInt(readByte(buf))
	r.Name = readString(buf)
	r.Map = readString(buf)
	r.Folder = readString(buf)
	r.Game = readString(buf)
	r.ID = toInt(readShort(buf))
	r.Players = toInt(readByte(buf))
	r.MaxPlayers = toInt(readByte(buf))
	r.Bots = toInt(readByte(buf))
	must(r.ServerType.UnmarshalBinary(readBytes(buf, 1)))
	must(r.Environment.UnmarshalBinary(readBytes(buf, 1)))
	must(r.Visibility.UnmarshalBinary(readBytes(buf, 1)))
	must(r.VAC.UnmarshalBinary(readBytes(buf, 1)))
	r.Version = readString(buf)

	// Check if EDF byte is present
	if buf.Len() < 1 {
		return nil
	}

	glog.V(2).Infof("steam: reading edf data (remaining bytes %v)", buf.Len())

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
