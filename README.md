<<<<<<< HEAD
# event-listener
k8s warning event listener 
=======
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o event-listener .
编写Dockerfile
docker build -t registry.cn-beijing.aliyuncs.com/kunpengcloud/event-listener:v3.3 .
docker push registry.cn-beijing.aliyuncs.com/kunpengcloud/event-listener:v3.3
>>>>>>> 753d53f (Initial commit)
