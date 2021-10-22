package dockerclient

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mikeletux/go-docker-manager/pkg/dockerclient/models"
	"github.com/mikeletux/go-docker-manager/pkg/httpclient"
)

// errors definition
var (
	ErrDockerInternalServerError = errors.New("there was an unknown error at docker daemon side")
	ErrImageDoesNotExist         = errors.New("the image selected does not exist")
	ErrContainerAlreadyExist     = errors.New("the container already exist")
	ErrContainerDoesNotExist     = errors.New("the container selected does not exist")
	ErrContainerIsRunning        = errors.New("cannot perform this operation with the container running")
	ErrContainerIsStopped        = errors.New("cannot perform this operation because the container is stopped")
	ErrExecInstanceDoesNotExist  = errors.New("the exec instance selected does not exist")
)

// SimpleDocker is a docker client that complies with the Docker interface
type SimpleDocker struct {
	// DockerEndpoint is the docker daemon endpoint to connect
	DockerEndpoint string

	// HttpClient is an struct that implements the interface HttpClient
	HttpClient httpclient.HttpClient
}

// NewSimpeDocker returns a SimpleDocker client given a docker endpoint and an HttpClient
func NewSimpeDocker(dockerEndpoint string, httpClient httpclient.HttpClient) *SimpleDocker {
	return &SimpleDocker{
		DockerEndpoint: dockerEndpoint,
		HttpClient:     httpClient,
	}
}

/* CheckIfImageAlreadyExists figures out if an image is already in the local repository.
Returns true if it is available in the local registry, false otherwise.*/
func (s *SimpleDocker) CheckIfImageAlreadyExists(dockerImage string, tag string) (bool, error) {
	httpResponse, err := s.HttpClient.Get(fmt.Sprintf("%s/images/%s:%s/json", s.DockerEndpoint, dockerImage, tag), nil)
	if err != nil {
		return false, fmt.Errorf("there was an issue with HTTP client when performing "+
			"GET on %s/images/%s:%s/json - %s", s.DockerEndpoint, dockerImage, tag, err)
	}

	switch httpResponse.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, ErrDockerInternalServerError
	}
}

// PullImageFromRegistry pulls an image from the docker registry given a docker Image name, image tag and image architecture
func (s *SimpleDocker) PullImageFromRegistry(dockerImage string, tag string, arch string) error {
	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/images/create?fromImage=%s&tag=%s&platform=%s",
		s.DockerEndpoint, dockerImage, tag, arch),
		nil, // No headers needed
		"")  // No body needed either
	if err != nil {
		return fmt.Errorf("there was an issue with HTTP client when performing POST "+
			"on %s/images/create?fromImage=%s&tag=%s&platform=%s - %s", s.DockerEndpoint, dockerImage, tag, arch, err)
	}

	switch httpResponse.StatusCode {
	case 200:
		return nil
	case 404:
		return ErrImageDoesNotExist
	default:
		return ErrDockerInternalServerError
	}
}

/* CreateContainer creates a a container given a container name, image name, image tag and list of commands for cmd.
It returns the ID of the new created container */
func (s *SimpleDocker) CreateContainer(containerName string, image string, tag string, cmd []string) (string, error) {
	httpRequestBody := models.CreateContainerBody{Cmd: cmd, Image: fmt.Sprintf("%s:%s", image, tag)}
	jsonBodyRequest, err := json.Marshal(httpRequestBody)
	if err != nil {
		return "", fmt.Errorf("json marshall issue when creating container - %s", err)
	}

	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/containers/create?name=%s", s.DockerEndpoint, containerName),
		map[string]string{"Content-Type": "application/json"},
		string(jsonBodyRequest))

	if err != nil {
		return "", fmt.Errorf("there was an issue with HTTP client performing POST on "+
			"%s/containers/create?name=%s - %s", s.DockerEndpoint, containerName, err)
	}

	switch httpResponse.StatusCode {
	case 201:
		var responseBody models.CreateContainerResponseBody
		err = json.Unmarshal(httpResponse.Body, &responseBody)
		if err != nil {
			return "", fmt.Errorf("json unmarshalling issue when creating container - %s", err)
		}
		return responseBody.ID, nil
	case 404:
		return "", ErrImageDoesNotExist
	case 409:
		return "", ErrContainerAlreadyExist
	default:
		return "", ErrDockerInternalServerError
	}
}

// RunContainer starts a new container given a container ID.
func (s *SimpleDocker) RunContainer(containerID string) error {
	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/containers/%s/start", s.DockerEndpoint, containerID),
		nil,
		"")
	if err != nil {
		return fmt.Errorf("there was an issue with HTTP client when performing POST on "+
			"%s/containers/%s/start - %s", s.DockerEndpoint, containerID, err)
	}

	switch httpResponse.StatusCode {
	case 204, 304:
		return nil
	case 404:
		return ErrContainerDoesNotExist
	default:
		return ErrDockerInternalServerError
	}

}

/* CheckIfContainerIsReady checks if a container is in running state
It returns true if it is running, false if in any other  */
func (s *SimpleDocker) CheckIfContainerIsReady(containerID string) (bool, error) {
	httpResponse, err := s.HttpClient.Get(fmt.Sprintf("%s/containers/%s/json", s.DockerEndpoint, containerID),
		nil)
	if err != nil {
		return false, fmt.Errorf("there was an issue with HTTP client when performing GET on "+
			"%s/containers/%s/json - %s", s.DockerEndpoint, containerID, err)
	}

	switch httpResponse.StatusCode {
	case 200:
		var containerStatus models.CheckContainerStatusBody
		err = json.Unmarshal(httpResponse.Body, &containerStatus)
		if err != nil {
			return false, fmt.Errorf("json unmarshalling issue when checking if container ready - %s", err)
		}

		if containerStatus.State.Status == "running" {
			return true, nil
		}
		return false, nil // The container is not yet in a running state

	case 404:
		return false, ErrContainerDoesNotExist
	default:
		return false, ErrDockerInternalServerError
	}
}

/* GenerateExecInstance generates a new exec instance on a container given a container ID and a command to run
It returns the exec ID */
func (s *SimpleDocker) GenerateExecInstance(containerID string, commands []string) (string, error) {
	httpRequestBody := models.GenerateExecInstanceBody{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          commands,
	}

	jsonBodyRequest, err := json.Marshal(httpRequestBody)
	if err != nil {
		return "", fmt.Errorf("json marshall issue when generating exec instance - %s", err)
	}

	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/containers/%s/exec", s.DockerEndpoint, containerID),
		map[string]string{"Content-Type": "application/json"},
		string(jsonBodyRequest))

	if err != nil {
		return "", fmt.Errorf("there was an issue with HTTP client when performing POST on "+
			"%s/containers/%s/exec - %s", s.DockerEndpoint, containerID, err)
	}

	switch httpResponse.StatusCode {
	case 201:
		var responseBody models.CreateExecResponseBody
		err = json.Unmarshal(httpResponse.Body, &responseBody)
		if err != nil {
			return "", fmt.Errorf("json unmarshalling issue when generating exec instance - %s", err)
		}
		return responseBody.ID, nil
	case 404:
		return "", ErrContainerDoesNotExist
	case 409:
		return "", ErrContainerIsStopped
	default:
		return "", ErrDockerInternalServerError
	}
}

/* StartExecInstance start an exec instance on a docker daemon given an exec ID.
It returns the stdout and stderr from inside the container */
func (s *SimpleDocker) StartExecInstance(execInstanceID string) (string, error) {
	httpRequestBody := models.StartExecInstance{
		Detach: false,
		Tty:    true,
	}

	jsonBodyRequest, err := json.Marshal(httpRequestBody)
	if err != nil {
		return "", fmt.Errorf("json marshalling issue when starting exec instance - %s", err)
	}

	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/exec/%s/start", s.DockerEndpoint, execInstanceID),
		map[string]string{"Content-Type": "application/json"},
		string(jsonBodyRequest))

	if err != nil {
		return "", fmt.Errorf("there was an issue with HTTP client when performing POST on "+
			"%s/exec/%s/start - %s", s.DockerEndpoint, execInstanceID, err)
	}

	switch httpResponse.StatusCode {
	case 200:
		return string(httpResponse.Body), nil
	case 404:
		return "", ErrExecInstanceDoesNotExist
	case 409:
		return "", ErrContainerIsStopped
	default:
		return "", ErrDockerInternalServerError
	}
}

/* StopContainer stops a container given a container ID.
Returns true if the container is stopped and false is the container was already stopped */
func (s *SimpleDocker) StopContainer(containerID string) (bool, error) {
	httpResponse, err := s.HttpClient.Post(fmt.Sprintf("%s/containers/%s/stop", s.DockerEndpoint, containerID),
		nil,
		"")
	if err != nil {
		return false, fmt.Errorf("there was an issue with HTTP client when performing POST on "+
			"%s/containers/%s/stop - %s", s.DockerEndpoint, containerID, err)
	}

	switch httpResponse.StatusCode {
	case 204:
		return true, nil
	case 304:
		return false, nil
	case 404:
		return false, ErrContainerDoesNotExist
	default:
		return false, ErrDockerInternalServerError
	}
}

// RemoveContainer removes a container given a container ID
func (s *SimpleDocker) RemoveContainer(containerID string) error {
	httpResponse, err := s.HttpClient.Delete(fmt.Sprintf("%s/containers/%s", s.DockerEndpoint, containerID),
		nil)

	if err != nil {
		return fmt.Errorf("there was an issue with HTTP client when performing DELETE on "+
			"%s/containers/%s - %s", s.DockerEndpoint, containerID, err)
	}

	switch httpResponse.StatusCode {
	case 204:
		return nil
	case 404:
		return ErrContainerDoesNotExist
	case 409:
		return ErrContainerIsRunning
	default:
		return ErrDockerInternalServerError
	}
}
