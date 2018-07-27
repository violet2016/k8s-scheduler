docker build -t hanacolor/hawq-kube-scheduler:1.0 .
docker build -t hanacolor/hawq-kube-scheduler-extender:1.0 -f ./Dockerfile.extender .
docker push hanacolor/hawq-kube-scheduler:1.0
docker push hanacolor/hawq-kube-scheduler-extender:1.0
