package hawq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	clientset "github.com/Pivotal-DataFabric/hawq-misc/pkg/client/clientset/versioned"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var cs *kubernetes.Clientset
var config *restclient.Config

const schedulerName = "hawq-scheduler"

// StartScheduler start hawq scheduler
func StartScheduler(stopCh <-chan struct{}) error {
	config, err := GetClusterConfig()
	if err != nil {
		return err
	}
	cs, err = kubernetes.NewForConfig(config)

	watch, err := cs.CoreV1().Pods("").Watch(metav1.ListOptions{FieldSelector: "spec.nodeName=,spec.schedulerName=" + schedulerName})
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

func schedule(pod *v1.Pod) error {
	nodesList, err := cs.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "!node-role.kubernetes.io/master",
	})
	if err != nil {
		return err
	}

	// Get all hawqclusters
	hawqclusterClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building hawqcluster clientset: %s", err.Error())
	}
	hawqclusters, err := hawqclusterClient.PivotaldataV1alpha1().HAWQClusters(pod.GetNamespace()).List(metav1.ListOptions{})
	if err != nil || hawqclusters == nil || len(hawqclusters.Items) == 0 {
		log.Println("not found any hawqcluster")
		return err
	}

	// Select all the unscheduled master/standby pods belong to the hawq clusters
	// but node name is not empty
	var names []string
	for _, cluster := range hawqclusters.Items {
		names = append(names, cluster.GetName())
	}
	// TODO will have bug if the cluster is terminating and try to start the same cluster again
	selector := fmt.Sprintf("app in (%s),role in (master, standby)", strings.Join(names, ","))
	podList, err := cs.CoreV1().Pods(pod.Namespace).List(metav1.ListOptions{
		LabelSelector: selector,
		FieldSelector: "spec.nodeName!=",
	})
	if err != nil {
		log.Fatal(err)
	}
	return findNode(pod, nodesList, podList)
}

func findNode(pod *v1.Pod, nodes *v1.NodeList, antiPods *v1.PodList) error {
	if len(nodes.Items) == 0 {
		return errors.New("There is no node available")
	}
	exclude := map[string]bool{}
	if len(antiPods.Items) > 0 {
		for _, anti := range antiPods.Items {
			exclude[anti.Spec.NodeName] = true
		}
	}
	// Get the first node that is not in exclude and assign to it
	for _, n := range nodes.Items {
		if n.Spec.Unschedulable {
			log.Printf("node %s is not schedulable, skip it.", n.Name)
			continue
		}
		if _, ok := exclude[n.Name]; !ok {
			err := assignToNode(pod, n)
			if err != nil {
				log.Printf("assign pod %s to node %s failed: %s\n", pod.Name, n.Name, err)
				continue
			} else {
				log.Printf("schedule pod %s to node %s\n", pod.Name, n.Name)
				return nil
			}
		}
	}
	return errors.New("No node is available for assignment")
}

func assignToNode(pod *v1.Pod, node v1.Node) error {
	binding := &v1.Binding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Binding",
		},
		ObjectMeta: metav1.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
			Namespace:  node.Namespace,
		},
	}

	err := cs.CoreV1().Pods(pod.Namespace).Bind(binding)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Successfully assigned %s to %s", pod.Name, node.Name)
	timestamp := time.Now().UTC()
	e := &v1.Event{
		Count:          1,
		Message:        message,
		ObjectMeta:     metav1.ObjectMeta{GenerateName: pod.Name + "-", Namespace: "default"},
		Reason:         "Scheduled",
		LastTimestamp:  metav1.Time{Time: timestamp},
		FirstTimestamp: metav1.Time{Time: timestamp},
		Type:           "Normal",
		Source:         v1.EventSource{Component: schedulerName},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      pod.Name,
			Namespace: "default",
			UID:       pod.UID,
		},
	}
	_, err = cs.CoreV1().Events("default").Create(e)
	if err != nil {
		log.Println("error in create event", err)
	}
	return nil
}

// SchedulePods will schedule the newly added pod
func SchedulePods(w watch.Interface, stopCh <-chan struct{}) {
	for {
		select {
		case event := <-w.ResultChan():
			if event.Type == watch.Added {
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					log.Println("receive non pod event")
					continue
				}
				log.Printf("Start to schedule pod %s in namespace %s\n", pod.Name, pod.Namespace)
				err := schedule(pod)
				if err != nil {
					log.Println("Error:", pod.Name, err)
				}
			}
		case <-stopCh:
			w.Stop()
			log.Println("stop watching")
			os.Exit(0)
		}
	}
}
