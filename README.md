# go-k8s-srv2

<h1 align="center">icap-service-2</h1>

<p align="center">
    <a href="https://github.com/k8-proxy/go-k8s-srv2/actions/workflows/build.yaml">
        <img src="https://github.com/k8-proxy/go-k8s-srv2/actions/workflows/build.yaml/badge.svg"/>
    </a>
    <a href="https://codecov.io/gh/k8-proxy/go-k8s-srv2">
        <img src="https://codecov.io/gh/k8-proxy/go-k8s-srv2/branch/main/graph/badge.svg"/>
    </a>	    
    <a href="https://goreportcard.com/report/github.com/k8-proxy/go-k8s-srv2">
      <img src="https://goreportcard.com/badge/k8-proxy/go-k8s-srv2" alt="Go Report Card">
    </a>
	<a href="https://github.com/k8-proxy/go-k8s-srv2/pulls">
        <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions welcome">
    </a>
    <a href="https://opensource.org/licenses/Apache-2.0">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache License, Version 2.0">
    </a>
    <a href="https://github.com/k8-proxy/go-k8s-srv2/releases/latest">
        <img src="https://img.shields.io/github/release/k8-proxy/go-k8s-srv1.svg?style=flat"/>
    </a>
</p>

This service is as per the architecture the first proxy service that will get the file from icap-server and upload it to minio

### Steps of processing
When it starts
- Listens on the queue
- Get the file from the processing outcome queue
- Copy that file to the shared disk
- Notify icap server on the adaptation outcome queue

## Configuration
- This pod need to mount the share storage mounted on icap server and that is how they will share the file together
- It's possible to have multiple replica of this service running. Only one will get the file and process itdj

### Docker build
- To build the docker image
```
git clone https://github.com/k8-proxy/go-k8s-srv2
cd k8-proxy/go-k8s-srv2
docker build -t <docker_image_name> .
```

- To run the container
First make sure that you have rabbitmq and minio running, then run the command bellow 

```
docker run -e ADAPTATION_REQUEST_QUEUE_HOSTNAME='<rabbit-host>' \ 
-e ADAPTATION_REQUEST_QUEUE_PORT='<rabbit-port>' \
-e MESSAGE_BROKER_USER='<rabbit-user>' \
-e MESSAGE_BROKER_PASSWORD='<rabbit-password>' \
-e MINIO_ENDPOINT='<minio-endpoint>' \ 
-e MINIO_ACCESS_KEY='<minio-access>' \ 
-e MINIO_SECRET_KEY='<minio-secret>' \ 
-e MINIO_SOURCE_BUCKET='<bucket-to-upload-file>' \ 
-e MINIO_CLEAN_BUCKET='<bucket-to-upload-file>' \
-e TRANSACTION_STORE_PATH='<path-to-report-file>' \
--name <docker_container_name> <docker_image_name>
```

# Testing steps

- Run the container as mentionned above

- Publish data reference to rabbitMq on queue name : adaptation-request-queue with the following data(table) :
* file-id : An ID for the file
* source-file-location : The full path to the file
* rebuilt-file-location : A full path representing the location where the rebuilt file will go to


- Check your container logs to see the processing

```
docker logs <container name>
```

# Rebuild flow to implement

![new-rebuild-flow-v2](https://github.com/k8-proxy/go-k8s-infra/raw/main/diagram/go-k8s-infra.png)
