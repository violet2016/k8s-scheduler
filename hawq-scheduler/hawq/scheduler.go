package hawq

import (
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// StartScheduler start hawq scheduler
func StartScheduler(stopCh <-chan struct{}) error {
	config, err := GetClusterConfig()
	if err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(config)

	watch, err := cs.CoreV1().Pods("").Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}
	SchedulePods(watch, stopCh)
	return nil
}

// GetClusterConfig from the env KUBECONFIG, default path ~/.kube/config
func GetClusterConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		// use the current context in kubeconfig
		// This is very useful for running locally.
		if home := os.Getenv("HOME"); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	log.Println("kubeconfig is", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err == nil {
		return config, err
	}
	return rest.InClusterConfig()
}

// SchedulePods will schedule the newly added pod
func SchedulePods(w watch.Interface, stopCh <-chan struct{}) {
	for {
		select {
		case event := <-w.ResultChan():
			log.Println(event)
			if event.Type == watch.Added {
				log.Println("ADDED")
			}
		case <-stopCh:
			w.Stop()
			log.Println("stop watching")
		}
	}
}
