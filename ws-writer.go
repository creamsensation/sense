package sense

import (
	"encoding/json"
	"slices"

	"github.com/creamsensation/auth"
	"github.com/creamsensation/socketer"
)

type WsWriter interface {
	Id(id ...int) WsWriter
	Session() WsWriter
	Json(value any) error
	Text(value string) error
}

type wsWriter struct {
	name string
	auth auth.Manager
	ids  []int
	ws   map[string]socketer.Ws
}

func createWsWriter(ws map[string]socketer.Ws, name string, auth auth.Manager) WsWriter {
	return &wsWriter{
		name: name,
		auth: auth,
		ids:  make([]int, 0),
		ws:   ws,
	}
}

func (s *wsWriter) Send() WsWriter {
	return s
}

func (s *wsWriter) Session() WsWriter {
	session := s.auth.Session().MustGet()
	if slices.Contains(s.ids, session.Id) {
		s.ids = append(s.ids, session.Id)
	}
	return s
}

func (s *wsWriter) Id(id ...int) WsWriter {
	for _, item := range id {
		if slices.Contains(s.ids, item) {
			s.ids = append(s.ids, item)
		}
	}
	return s
}

func (s *wsWriter) Json(value any) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(s.ids) == 0 {
		s.ws[s.name].Broadcast(bytes)
		return nil
	}
	clients, err := s.ws[s.name].Find(s.ids...)
	if err != nil {
		return err
	}
	for _, c := range clients {
		if err := c.Write(bytes); err != nil {
			return err
		}
	}
	return nil
}

func (s *wsWriter) Text(value string) error {
	if len(s.ids) == 0 {
		s.ws[s.name].Broadcast([]byte(value))
		return nil
	}
	clients, err := s.ws[s.name].Find(s.ids...)
	if err != nil {
		return err
	}
	for _, c := range clients {
		if err := c.Write([]byte(value)); err != nil {
			return err
		}
	}
	return nil
}
