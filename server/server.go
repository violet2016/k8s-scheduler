package server

import (
	"encoding/json"
	"log"
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
	log.Println("head is", head)
	if head != prioritize && head != filter {
		log.Println("unknown method")
		http.Error(res, "Unknown method", http.StatusNotFound)
		return
	}
	decoder := json.NewDecoder(req.Body)
	var args schedulerapi.ExtenderArgs
	if err := decoder.Decode(&args); err != nil {
		log.Println("decode error")
		http.Error(res, "Decode error", http.StatusBadRequest)
		return
	}
	var result *schedulerapi.ExtenderFilterResult
	switch head {
	case filter:
		log.Println("get filter! nodes number", len(args.Nodes.Items), args.Pod.Name)
		result = s.filterHandler.Filter(args.Pod, args.Nodes)

	case prioritize:
		log.Println("get prioritize", args.NodeNames, args.Pod.Name)
	}
	if result != nil {
		if resultBody, err := json.Marshal(result); err != nil {
			panic(err)
		} else {
			log.Println("finished filter for pod", args.Pod.Name)
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusOK)
			res.Write(resultBody)
			return
		}
	}
	http.Error(res, "Not Found", http.StatusNotFound)
}
