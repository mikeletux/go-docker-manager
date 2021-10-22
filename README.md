# Go-docker-manager
This project consist of a Go app that connects to a Docker backend, spans a Ubuntu container and shows live CPU/Memory information from inside the container. Also, as per user interaction, the container can be destroyed on demand.

## Prerequisites
In order to run this project, the following prerequisites need to be met:
  - **Operating system**: Ubuntu 20.04
  - **Docker**: version 2.10.8, build 3967b7d (This version implements Docker Engine API v.1.41)

(These are the requirements for the Docker backend instance. Other OSs and Docker versions might work, but no testing has been done)

To build this project manually, the following dependencies are needed:
  - Go 1.16

To run this project as a Docker container, please refer to *Running Go-docker-manager as a container* section.

## Changes needed in Docker backend to access the Rest API
By default, Docker host does not expose the HTTP Rest API for consumption, it only does it locally through UNIX sockets.
To enable the HTTP Rest API, perform the following steps:
  - Open up the file at */lib/systemd/system/docker.service*
  - Find the line that starts with *ExecStart* and add the following string (feel free to use a different TCP port):
  ```
  -H=tcp://0.0.0.0:2375
  ```
  - Reload the Docker daemon:
  ```
  systemctl daemon-reload
  ```
  - Restart the Docker daemon:
  ```
  systemctl restart docker
  ```
After this, you will be able to use this adapter against your Docker backend.
  
This project leverages on HTTP rather than UNIX sockets for accessing the rest API. This is because through HTTP users can control both local or remote Docker backends.

## Adapter configuration
When compiled into a binary, some configuration needs to be placed before executing. This configuration needs to be passed either using *flags* or *OS environment variablers*. **Also notice that if both are set, environment variables have higher priority.**  
The only input the user needs to pass is the endpoint from the Rest API the Docker backend is listening to. This can be done using the **-e** flag or the **DOCKER_MANAGER_ENDPOINT** env var:
  - Flags:
  ```
  ./dockermanager -e http://192.168.1.150:2375
  ```
  - Environment var:
  ```
  export DOCKER_MANAGER_ENDPOINT=http://192.168.1.150:2375
  ./dockermanager
  ```

## Using the application
Right after the application is executed, if everything is ok, it will start an Ubuntu 20.04 container and it will execute periodically a command that outputs statistics about CPU and memory. The program can be finished typing the character *e* and pressing *ENTER*. After that, the container will be stopped and destroyed.  

![](./images/demo.gif)

## Running Go-docker-manager as a container
The project comes alongside a Dockerfile that can be used to build a Docker image with the project embedded with all its dependencies. To build the image use the below command:
```
docker build -t dockermanager .
```
To run the container, some configuration needs to be injected using environment variables. Please refer to the *Adapter configuration* to more information about it. Use the example below as reference:
```
docker run -it \
-e DOCKER_MANAGER_ENDPOINT=http://192.168.1.150:2375 \
--name my-docker-manager \
dockermanager
```

## Application flows
To give the user some insight about how the application works, here are the workflows that the Go app follows:
  - During application start up:
    - A Docker client is created to speak with the Docker backend
    - The program checks if the Ubuntu 20.04 image already exists, if not it downloads it from Dockerhub
    - From the above image, a container is created and initiated
    - The program waits until the container is ready. If after 180 seconds is not, it fails

  - Application lifecycle:
    - The app executes commands into the container that retrieve CPU and memory statistics
    - It prints by stdout the result
    - Perform the two steps above indefinitely until a user type the character *e* and press *ENTER*
    - When the step above is done, the program shuts down the container and removes it from the Docker backend

## Acceptance testing
The project also comes with some tests to check that the implementation of the Docker client does what it is supposed to be built for.  

The tests can be found in folder *pkg/dockerclient* under the file *simpleDocker_test.go*. These are integration tests, so they run against real infrastructure. Please make sure you have a Docker backend configured appropriately (Check *Changes needed in Docker backend to access the Rest API* section for that).  

Before executing, in the tests file change the constant *DockerEndpoint* to your Docker backened address. After doing that, you're set to run the tests. Please use the below command for that:
```
go test -p 1 .
```

/Miguel Sama 2021