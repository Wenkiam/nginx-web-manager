package web

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/context"
	gsessions "github.com/gorilla/sessions"
	"log"
	"net/http"
)

const (
	errorFormat = "[sessions] ERROR! %s\n"
)
const flashesKey = "_flash"

func setupSession() {
	engine.Use(Sessions("session"))
}

func Sessions(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := &session{name, c.Request, make(map[interface{}]*gsessions.Session), false, c.Writer}
		c.Set(sessions.DefaultKey, s)
		defer context.Clear(c.Request)
		c.Next()
	}
}

type session struct {
	name       string
	request    *http.Request
	sessionMap map[interface{}]*gsessions.Session
	written    bool
	writer     http.ResponseWriter
}

func (s *session) ID() string {
	return s.name
}

func (s *session) Get(key interface{}) interface{} {
	return s.Session(key).Values[key]
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Session(key).Values[key] = val
	s.written = true
}

func (s *session) Delete(key interface{}) {
	delete(s.Session(key).Values, key)
	s.written = true
}

func (s *session) Clear() {
	for key := range s.sessionMap {
		s.Delete(key)
	}
}

func (s *session) AddFlash(value interface{}, vars ...string) {
	key := flashesKey
	if len(vars) > 0 {
		key = vars[0]
	}
	var flashes []interface{}
	if v, ok := s.sessionMap[key]; ok {
		flashes = v.Values[key].([]interface{})
	}
	s.Session(key).Values[key] = append(flashes, value)
	s.written = true
}

func (s *session) Flashes(vars ...string) []interface{} {
	s.written = true
	var flashes []interface{}
	key := flashesKey
	if len(vars) > 0 {
		key = vars[0]
	}
	if v, ok := s.Session(key).Values[key]; ok {
		// Drop the flashes and return it.
		delete(s.Session(key).Values, key)
		flashes = v.([]interface{})
	}
	return flashes
}

func (s *session) Options(options sessions.Options) {
	s.written = true
	for _, v := range s.sessionMap {
		v.Options = options.ToGorillaOptions()
	}
}

func (s *session) Save() error {
	if s.Written() {
		var e error
		for _, v := range s.sessionMap {
			e = v.Save(s.request, s.writer)
			if e != nil {
				return e
			}
		}
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

func (s *session) Written() bool {
	return s.written
}

func (s *session) Session(key interface{}) *gsessions.Session {
	ss := s.sessionMap[key]
	if ss != nil {
		return ss
	}
	secret, ok := key.(string)
	if !ok || secret == "" {
		secret = "nwm-secret"
	}
	var err error
	cs := cookie.NewStore([]byte(secret))
	ss, err = cs.Get(s.request, fmt.Sprintf("%v", key))
	if err != nil {
		log.Printf(errorFormat, err)
	}
	s.sessionMap[key] = ss
	return ss
}
