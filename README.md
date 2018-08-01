# k8s-scheduler

This repo contains 2 implementations of k8s scheduler examples. The rules of scheduling pods is quite simple, but they can show you how to create your own schedulers in k8s.

1. A standalone scheduler server: [hawq-scheduler](#hawq-scheduler)
2. An example of second scheduler in k8s 1.11, using extenders: [hawq-scheduler-extender](#hawq-scheduler-extender)

## Hawq scheduler

Hawq scheduler is a server that can run outside of k8s or in a container in k8s. 

Hawq scheduler use client-go to get the pod that use **hawq-scheduler** as it's schedulerName. Here is an example for creating a pod using hawq-scheduler:

```
apiVersion: v1
kind: Pod
metadata:
  name: hawq-cluster-1-master-0
spec:
  containers:
    - name: hawq-master
      image: hawqbeijing/hawq_proxy:vcheng
  schedulerName: hawq-scheduler
```

With this type of scheduler, k8s has no idea if the scheduler is running or not, or even does not know it exist.



## Hawq scheduler extender - Kubernetes scheduler extender

hawq scheduler extender can run with a registered k8s scheduler, it is set with a config file. Since k8s has moved scheduler extender from plugin to cmd recently, there is few examples on how to create a scheduler with extender. Well, this repo is what you can refer to :)

[Read this link for more details](./hawq-scheduler-extender/README.md)
