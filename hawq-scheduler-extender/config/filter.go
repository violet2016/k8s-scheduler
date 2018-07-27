package config

import "k8s.io/api/core/v1"

// HawqFilter contains the special rules to filter node for hawq
func HawqFilter(pod *v1.Pod, node *v1.Node) (bool, error) {
	// filter out master
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		return false, nil
	}
	return true, nil
}
