package server

import (
	"encoding/json"
	"net/http"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// Server handle the api reqeust routing
type Server struct {
	filterHandler *FilterHandler
}
type filterFunc func(*v1.Pod, *v1.Node) (bool, error)

// NewServer create a server
func NewServer(f filterFunc) *Server {
	return &Server{filterHandler: &FilterHandler{FilterOneNode: f}}
}
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = SplitPath(req.URL.Path)
	if head != prioritize && head != filter {
		http.Error(res, "Unknown method", http.StatusNotFound)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var args schedulerapi.ExtenderArgs
	if err := decoder.Decode(&args); err != nil {
		http.Error(res, "Decode error", http.StatusBadRequest)
		return
	}
	switch head {
	case filter:
		return
	case prioritize:
	}

	http.Error(res, "Not Found", http.StatusNotFound)
}
