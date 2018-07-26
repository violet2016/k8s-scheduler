package main

import (
	"net/http"

	"github.com/violet2016/k8s-scheduler/hawq-scheduler-extender/config"
	"github.com/violet2016/k8s-scheduler/server"
)

func main() {
	s := server.NewServer(config.HawqFilter)
	http.ListenAndServe(":8000", s)
}
