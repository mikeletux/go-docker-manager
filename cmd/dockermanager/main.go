package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mikeletux/go-docker-manager/pkg/dockerclient"
	"github.com/mikeletux/go-docker-manager/pkg/httpclient"

	"github.com/gosuri/uilive" // library for updating terminal in real time :)
)

const (
	DockerImage         = "ubuntu"
	DockerImageTag      = "20.04"
	DockerImageArch     = "x86-64"
	DockerContainerName = "ubuntu2004"

	DefaultDockerEndpoint = "http://localhost:2375"
)

var (
	printHelp      bool
	dockerEndpoint string
)

func init() {
	flag.BoolVar(&printHelp, "h", false, "shows help")
	flag.StringVar(&dockerEndpoint, "e", DefaultDockerEndpoint, "docker endpoint to connect")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if printHelp {
		flag.Usage()
		return
	}
	// ENV vars have higher priority than flags if set
	envVar, present := os.LookupEnv("DOCKER_MANAGER_ENDPOINT")
	if present {
		dockerEndpoint = envVar
	}

	simpleHttpClient := httpclient.NewSimpleHttpClient()

	log.Printf("docker manager set to %s", dockerEndpoint)
	dockerClient := dockerclient.NewSimpeDocker(dockerEndpoint, simpleHttpClient)

	exists, err := dockerClient.CheckIfImageAlreadyExists(DockerImage, DockerImageTag)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		log.Printf("couldn't find image %s:%s locally, downloading...", DockerImage, DockerImageTag)
		err = dockerClient.PullImageFromRegistry(DockerImage, DockerImageTag, DockerImageArch)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("initiating container %s from image %s:%s", DockerContainerName, DockerImage, DockerImageTag)
	containerID, err := dockerClient.CreateContainer(DockerContainerName, DockerImage, DockerImageTag, []string{"sleep", "infinity"})
	if err != nil {
		log.Fatal(err)
	}

	err = dockerClient.RunContainer(containerID)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("waiting for container to be in running state...")
	timeoutTicket := time.NewTicker(180 * time.Second)
	checkTicket := time.NewTicker(time.Second)
out:
	for {
		select {
		case <-timeoutTicket.C:
			log.Fatal("the container didn't get into running status for 180s")
		case <-checkTicket.C:
			status, err := dockerClient.CheckIfContainerIsReady(containerID)
			if err != nil {
				log.Fatal(err)
			}
			if status {
				break out
			}
		}
	}

	var wg sync.WaitGroup

	done := make(chan bool) // This channel allow communication between executeCommandAndPrint and readKeyboardEvent go routines

	// go routine that outputs the container commands
	wg.Add(1)
	go executeCommandAndPrint(&wg, dockerClient, containerID, done)

	// go routine that listens to keyboard event to finish
	wg.Add(1)
	go readKeyboardEvent(&wg, done)

	wg.Wait()

	log.Println("Stopping the container, please wait...")
	_, err = dockerClient.StopContainer(containerID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Removing the container, please wait...")
	err = dockerClient.RemoveContainer(containerID)
	if err != nil {
		log.Fatal(err)
	}
}

func executeCommandAndPrint(wg *sync.WaitGroup, dockerClient dockerclient.Docker, containerID string, done <-chan bool) {
	defer wg.Done()

	ticker := time.NewTicker(800 * time.Millisecond)
	writer := uilive.New()
	writer.Start()

	for {
		select {
		case <-done:
			writer.Stop()
			return
		case <-ticker.C:
			execID, err := dockerClient.GenerateExecInstance(containerID, []string{"/bin/sh", "-c", "top -b -n 1 | head -4 | tail -2"})
			if err != nil {
				log.Fatal(err)
			}

			output, err := dockerClient.StartExecInstance(execID)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintf(writer, "Type \"e\" and press ENTER to finish\n%s", output)
		}
	}
}

func readKeyboardEvent(wg *sync.WaitGroup, done chan bool) {
	defer wg.Done()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "e" {
			done <- true
			return
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Go Docker Manager v0.1.0
Usage: dockermanager [-e endpoint]

Options:
`)
	flag.PrintDefaults()
}
