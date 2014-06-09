package steam

import (
	"fmt"
	"github.com/golang/glog"
	"errors"
	"bytes"
)

type Player struct{
	index byte
	name string
	score int32
	duration float32
}

type A2SPlayersResponse struct{
	playersCount byte
	players []Player
}

func (a *A2SPlayersResponse) UnMarshalBinary(data []byte) (err error) {
	glog.V(2).Infof("unmarshalling binary for A2SPlayersResponse: %v", data)
	buf := bytes.NewBuffer(data)

	if header := readByte(buf); header != 0x44{
		return errors.New("steam: invalid header in the a2splayersresponse")
	}

	a.playersCount = readByte(buf)
	a.players = make([]Player, a.playersCount)

	for i := 0; i < int(a.playersCount); i++ {
		p := &a.players[i]
		p.index = readByte(buf)
		p.name = readString(buf)
		p.score = readLong(buf)
		p.duration = readFloat(buf)
	}

	return nil
}

func (a *A2SPlayersResponse) String() string{
	str := fmt.Sprintf("players count: %v\n\n", a.playersCount)	

	for _, player := range a.players {
		str += fmt.Sprintf("player index: %v\n", player.index)	
		str += fmt.Sprintf("player name: %v\n", player.name)	
		str += fmt.Sprintf("player score: %v\n", player.score)	
		str += fmt.Sprintf("player duration: %v seconds\n\n", player.duration)	
	}

	return str
}