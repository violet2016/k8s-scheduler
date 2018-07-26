package server

import (
	"encoding/json"
	"log"
	"net/http"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// FilterHandler handle
type FilterHandler struct {
	FilterOneNode filterFunc
}

// Handle process the url and call Filter function
func (f *FilterHandler) Handle(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	var args schedulerapi.ExtenderArgs
	encoder := json.NewEncoder(res)
	if err := decoder.Decode(&args); err != nil {
		http.Error(res, "Decode error", http.StatusBadRequest)
		return
	}

	resp := f.Filter(args.Pod, args.Nodes)

	if err := encoder.Encode(resp); err != nil {
		log.Fatalf("Failed to encode %+v", resp)
	}
}

// Filter do the node filter
func (f *FilterHandler) Filter(pod *v1.Pod, nodes *v1.NodeList) *schedulerapi.ExtenderFilterResult {
	canSchedule := make([]v1.Node, 0, len(nodes.Items))
	canNotSchedule := make(map[string]string)

	for _, node := range nodes.Items {
		result, err := f.FilterOneNode(pod, &node)
		if err != nil {
			canNotSchedule[node.Name] = err.Error()
		} else {
			if result {
				canSchedule = append(canSchedule, node)
			}
		}
	}

	result := &schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{

			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	return result
}
