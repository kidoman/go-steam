package steam

import (
	"bytes"
	"io"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

type rconSocket struct {
	conn net.Conn
}

func newRCONSocket(dial DialFn, addr string) (*rconSocket, error) {
	conn, err := dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &rconSocket{conn}, nil
}

func (s *rconSocket) close() {
	s.conn.Close()
}

func (s *rconSocket) send(p []byte) error {
	if err := s.conn.SetWriteDeadline(time.Now().Add(400 * time.Millisecond)); err != nil {
		return err
	}

	_, err := s.conn.Write(p)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("steam: could not write rcon request")

		return err
	}

	return nil
}

func (s *rconSocket) receive() (_ []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	if err := s.conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tr := io.TeeReader(s.conn, &buf)

	total := int(readLong(tr))
	log.WithFields(logrus.Fields{
		"total": total + 4,
	}).Debug("steam: reading rcon packet")

	for total > 0 {
		log.WithFields(logrus.Fields{
			"bytes": total,
		}).Debug("steam: reading rcon response")

		b := make([]byte, total)
		n, err := s.conn.Read(b)
		if n > 0 {
			log.WithFields(logrus.Fields{
				"bytes": n,
			}).Debug("steam: read rcon response")

			_, err := buf.Write(b)
			if err != nil {
				return nil, err
			}

			total -= n
		}
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Error("steam: could not receive data")

			if err == io.EOF {
				return nil, err
			}
		}

		log.WithFields(logrus.Fields{
			"bytes": total,
		}).Debug("steam: remaining rcon response")
	}

	log.WithFields(logrus.Fields{
		"size": buf.Len(),
	}).Debug("steam: read rcon packet")

	return buf.Bytes(), nil
}
