This is an example of how to run a second k8s scheduler based on k8s offical documents.

First run
```
kubectl create -f example-scheduler.yaml

kubectl edit clusterrole system:kube-scheduler
```

Then add **my-scheduler** in
```
resourceNames:
    - kube-scheduler
    - my-scheduler #add here
```

kubectl create -f example-pod.yaml

You should see an event the the pod is scheduled by my-scheduler in k8s events
```
kubectl get events
```