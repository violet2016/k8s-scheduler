package hawq

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	clientset "github.com/Pivotal-DataFabric/hawq-misc/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var cs *kubernetes.Clientset
var config *restclient.Config

// StartScheduler start hawq scheduler
func StartScheduler(stopCh <-chan struct{}) error {
	config, err := GetClusterConfig()
	if err != nil {
		return err
	}
	cs, err = kubernetes.NewForConfig(config)

	watch, err := cs.CoreV1().Pods("").Watch(metav1.ListOptions{FieldSelector: "spec.nodeName="})
	if err != nil {
		return err
	}
	SchedulePods(watch, stopCh)
	return nil
}

// GetClusterConfig from the env KUBECONFIG, default path ~/.kube/config
func GetClusterConfig() (*restclient.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		// use the current context in kubeconfig
		// This is very useful for running locally.
		if home := os.Getenv("HOME"); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	log.Println("kubeconfig is", kubeconfig)
	var err error
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err == nil {
		return config, err
	}
	return restclient.InClusterConfig()
}

func schedule() error {
	// nodesList, err := cs.CoreV1().Nodes().List(metav1.ListOptions{})
	// if err != nil {
	// 	return err
	// }

	// Get all hawqclusters
	hawqclusterClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building hawqcluster clientset: %s", err.Error())
	}
	hawqclusters, err := hawqclusterClient.PivotaldataV1alpha1().HAWQClusters("").List(metav1.ListOptions{})
	if err != nil || hawqclusters == nil {
		log.Println("not found any hawqcluster")
	}

	// Select all the pods belong to the hawq clusters
	var names []string
	for _, cluster := range hawqclusters.Items {
		names = append(names, cluster.GetClusterName())
	}
	selector := fmt.Sprintf("app in (%s),role in (master, standby)", strings.Join(names, ","))
	podList, err := cs.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: selector})
	for _, p := range podList.Items {
		log.Println("pod name is", p.Name)
	}
	return nil
}

// SchedulePods will schedule the newly added pod
func SchedulePods(w watch.Interface, stopCh <-chan struct{}) {
	for {
		select {
		case event := <-w.ResultChan():
			if event.Type == watch.Added {
				schedule()
			}
		case <-stopCh:
			w.Stop()
			log.Println("stop watching")
			os.Exit(0)
		}
	}
}
