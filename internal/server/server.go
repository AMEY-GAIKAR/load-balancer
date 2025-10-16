package server

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Server struct {
	Url          *url.URL
	Weight       int
	ReverseProxy *httputil.ReverseProxy
	Connections  int64
	alive        bool
	mutex        sync.RWMutex
}

func (s *Server) SetStatus(status bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.alive = status
}

func (s *Server) IsAlive() bool {
	var alive bool
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	alive = s.alive
	return alive
}

func (s *Server) SetWeight(weight int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Weight = weight
}
