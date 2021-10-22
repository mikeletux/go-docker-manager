package dockerclient

import (
	"testing"

	"github.com/mikeletux/go-docker-manager/pkg/httpclient"
)

const DockerEndpoint = "http://localhost:2375"

func TestSimpleDocker_CheckIfImageAlreadyExists(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to make sure that exists
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Test
	type args struct {
		dockerImage string
		tag         string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Check that a image that exist",
			args: args{
				dockerImage: "ubuntu",
				tag:         "20.04",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Check a image that doesn't exist",
			args: args{
				dockerImage: "centos",
				tag:         "7",
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerClient.CheckIfImageAlreadyExists(tt.args.dockerImage, tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.CheckIfImageAlreadyExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleDocker.CheckIfImageAlreadyExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleDocker_PullImageFromRegistry(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	type args struct {
		dockerImage string
		tag         string
		arch        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Pulling an image that actually exists",
			args: args{
				dockerImage: "ubuntu",
				tag:         "20.04",
				arch:        "x86-64",
			},
			wantErr: false,
		},
		{
			name: "Pulling an image that doesn't exists",
			args: args{
				dockerImage: "ubuntu",
				tag:         "20.07", // This tag doesn't exist in dockerhub
				arch:        "x86-64",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dockerClient.PullImageFromRegistry(tt.args.dockerImage, tt.args.tag, tt.args.arch); (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.PullImageFromRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleDocker_CreateContainer(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		containerName string
		image         string
		tag           string
		cmd           []string
	}
	tests := []struct {
		name string
		args args
		// want    string // We cannot know the container ID upfront
		wantErr bool
	}{
		{
			name: "Create a container with a image that exists",
			args: args{
				containerName: "ubuntu2004",
				image:         "ubuntu",
				tag:           "20.04",
				cmd:           []string{"free", "-h"},
			},
			wantErr: false,
		},
		{
			name: "Create a container that already exists",
			args: args{
				containerName: "ubuntu2004",
				image:         "ubuntu",
				tag:           "20.04",
				cmd:           []string{"free", "-h"},
			},
			wantErr: true,
		},
		{
			name: "Try to create a container with an image that doesn't exist",
			args: args{
				containerName: "centos10",
				image:         "centos",
				tag:           "10.0",
				cmd:           []string{"free", "-h"},
			},
			wantErr: true,
		},
	}
	var containerID string // We keep the container reference for later removal

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := dockerClient.CreateContainer(tt.args.containerName, tt.args.image, tt.args.tag, tt.args.cmd)
			if len(got) > 0 {
				containerID = got
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.CreateContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) == 0 { // If the len of the ID is > 0, we can consider that it returns an ID
				t.Errorf("SimpleDocker.CreateContainer() = %v, want len(ID) > 0 ", got)
			}
		})
	}

	// Remove container
	err = dockerClient.RemoveContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_RunContainer(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	id, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"free", "-h"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		containerID string
		wantErr     bool
	}{
		{
			name:        "Run a container that exist",
			containerID: id,
			wantErr:     false,
		},
		{
			name:        "Try to run the same container again",
			containerID: id,
			wantErr:     false,
		},
		{
			name:        "Try to run a container with an ID that doesn't exist",
			containerID: "fakefakefakefakefake",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dockerClient.RunContainer(tt.containerID); (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.RunContainer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Stop container
	_, err = dockerClient.StopContainer(id)
	if err != nil {
		t.Fatal(err)
	}

	// Remove container
	err = dockerClient.RemoveContainer(id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_CheckIfContainerIsReady(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	id, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"sleep", "infinity"})
	if err != nil {
		t.Fatal(err)
	}

	// Run container
	err = dockerClient.RunContainer(id)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		containerID string
		isRunning   bool // this will flag if the container must be running or stopped
		want        bool
		wantErr     bool
	}{
		{
			name:        "Check a running container",
			containerID: id,
			isRunning:   true,
			want:        true,
			wantErr:     false,
		},
		{
			name:        "Check a stopped container",
			containerID: id,
			isRunning:   false,
			want:        false,
			wantErr:     false,
		},
		{
			name:        "Check a container that doesn't exist",
			containerID: "fakefakefakefake",
			isRunning:   true,
			want:        false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		if !tt.isRunning {
			// Stop container
			_, err = dockerClient.StopContainer(id)
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerClient.CheckIfContainerIsReady(tt.containerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.CheckIfContainerIsReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleDocker.CheckIfContainerIsReady() = %v, want %v", got, tt.want)
			}
		})
	}

	// Remove container
	err = dockerClient.RemoveContainer(id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_GenerateExecInstance(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	id, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"sleep", "infinity"})
	if err != nil {
		t.Fatal(err)
	}

	// Run container
	err = dockerClient.RunContainer(id)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		containerID string
		commands    []string
		isRunning   bool // This flag allows to keep running or stopping the container for testing purposes
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create exec instance in a running container",
			args: args{
				containerID: id,
				commands:    []string{"free", "-h"},
				isRunning:   true,
			},
			wantErr: false,
		},
		{
			name: "Try to create an exec instance in a stopped container",
			args: args{
				containerID: id,
				commands:    []string{"free", "-h"},
				isRunning:   false,
			},
			wantErr: true,
		},
		{
			name: "Try to create an exec instance in a non existent container",
			args: args{
				containerID: "fakefakefakefake",
				commands:    []string{"free", "-h"},
				isRunning:   true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if !tt.args.isRunning {
			// Stop container
			_, err = dockerClient.StopContainer(id)
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerClient.GenerateExecInstance(tt.args.containerID, tt.args.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.GenerateExecInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) == 0 {
				t.Errorf("SimpleDocker.GenerateExecInstance() = %v, want len(ID) > 0 ", got)
			}
		})
	}

	// Remove container
	err = dockerClient.RemoveContainer(id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_StartExecInstance(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	containerID, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"sleep", "infinity"})
	if err != nil {
		t.Fatal(err)
	}

	// Run container
	err = dockerClient.RunContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}

	// Create exec instance
	execId, err := dockerClient.GenerateExecInstance(containerID, []string{"uname"})
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		execInstanceID     string
		isContainerRunning bool // This flag allows to keep running or stopping the container for testing purposes
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Start a exec from an existing exec instace in a running container",
			args: args{
				execInstanceID:     execId,
				isContainerRunning: true,
			},
			want:    "Linux\r\n",
			wantErr: false,
		},
		{
			name: "Start a exec from an existing exec instace in a stopped container",
			args: args{
				isContainerRunning: false,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Start a exec from an non existing exec instace",
			args: args{
				execInstanceID:     "fakefakefakefake",
				isContainerRunning: true,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if !tt.args.isContainerRunning {
			// Create new exec instance
			execId, err := dockerClient.GenerateExecInstance(containerID, []string{"uname"})
			if err != nil {
				t.Fatal(err)
			}
			tt.args.execInstanceID = execId

			// Stop container
			_, err = dockerClient.StopContainer(containerID)
			if err != nil {
				t.Fatal(err)
			}

		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerClient.StartExecInstance(tt.args.execInstanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.StartExecInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleDocker.StartExecInstance() = %v, want %v", got, tt.want)
			}
		})
	}

	// Remove container
	err = dockerClient.RemoveContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_StopContainer(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	containerID, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"sleep", "infinity"})
	if err != nil {
		t.Fatal(err)
	}

	// Run container
	err = dockerClient.RunContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		containerID string
		want        bool
		wantErr     bool
	}{
		{
			name:        "Stop a running container",
			containerID: containerID,
			want:        true,
			wantErr:     false,
		},
		{
			name:        "Try to stop a stopped container",
			containerID: containerID,
			want:        false,
			wantErr:     false,
		},
		{
			name:        "Try to stop a non existing container",
			containerID: "fakefakefakefake",
			want:        false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerClient.StopContainer(tt.containerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.StopContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleDocker.StopContainer() = %v, want %v", got, tt.want)
			}
		})
	}

	// Remove container
	err = dockerClient.RemoveContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleDocker_RemoveContainer(t *testing.T) {
	// Create HTTP client
	httpClient := httpclient.NewSimpleHttpClient()

	// Create Docker client
	dockerClient := NewSimpeDocker(DockerEndpoint, httpClient)

	// Download an image to create a container from
	err := dockerClient.PullImageFromRegistry("ubuntu", "20.04", "x86-64")
	if err != nil {
		t.Fatal(err)
	}

	// Create container
	containerID, err := dockerClient.CreateContainer("ubuntu2004", "ubuntu", "20.04", []string{"sleep", "infinity"})
	if err != nil {
		t.Fatal(err)
	}

	// Run container
	err = dockerClient.RunContainer(containerID)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		containerID        string
		isContainerRunning bool // This flag allows to keep running or stopping the container for testing purposes
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Try to remove a running container",
			args: args{
				containerID:        containerID,
				isContainerRunning: true,
			},
			wantErr: true,
		},
		{
			name: "Remove a stopped container",
			args: args{
				containerID:        containerID,
				isContainerRunning: false,
			},
			wantErr: false,
		},
		{
			name: "Try to remove a non existing container",
			args: args{
				containerID:        "fakefakefakefake",
				isContainerRunning: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if !tt.args.isContainerRunning {
			// Stop container
			_, err = dockerClient.StopContainer(containerID)
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := dockerClient.RemoveContainer(tt.args.containerID); (err != nil) != tt.wantErr {
				t.Errorf("SimpleDocker.RemoveContainer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
