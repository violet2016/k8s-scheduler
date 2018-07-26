package config

import "k8s.io/api/core/v1"

func HawqFilter(pod *v1.Pod, node *v1.Node) (bool, error) {
	return true, nil
}
