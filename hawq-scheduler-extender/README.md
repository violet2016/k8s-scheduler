# How to create a second k8s scheduler and an extender

## Create second k8s scheduler

The [k8s documents](https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/) gives a good example to deploy another k8s scheduler. You can also find the useful yaml and commands in [artifacts/example](./artifacts/example).

What you need to notice is that:

1. In the ClusterRoleBinding part of yaml:
``` 
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: my-scheduler-as-kube-scheduler
subjects:
- kind: ServiceAccount
  name: my-scheduler
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: kube-scheduler # should be system:kube-scheduler
  apiGroup: rbac.authorization.k8s.io
```
The name of roleRef should be system:kube-scheduler in 1.11. Otherwise it will report an error that kube-scheduler does not exist.

And, even you change it to system:kube-scheduler, it still doesn't have privilege to obtail storageclass, with is needed in scheduler. Change it to **cluster-admin** will fix this but may have security issue.

## Add extender to scheduler

You can config your scheduler with a file option **--config**.

The config file example is [here](./hack/docker/config.yaml). The extender is set in this part:
```
algorithmSource:
  policy:
    file:
      path: /opt/policy.json
```
The file content example is [here](./hack/docker/policy.json). It is a json file that contains predicates, priorities and extenders.