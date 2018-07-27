kubectl create -f example-scheduler.yaml

kubectl edit clusterrole system:kube-scheduler

add "my-scheduler" in
```
resourceNames:
    - kube-scheduler
    - my-scheduler #add here
```

kubectl create -f example-pod.yaml