package main

import (
	"log"

	"github.com/Pivotal-DataFabric/hawq-misc/pkg/signals"
	"github.com/violet2016/k8s-scheduler/hawq-scheduler/hawq"
)

func main() {
	stopCh := signals.SetupSignalHandler()
	err := hawq.StartScheduler(stopCh)
	if err != nil {
		log.Println(err)
	}
}
