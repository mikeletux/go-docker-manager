package dockerclient

// Docker is the interface any docker client must comply with
type Docker interface {
	/* CheckIfImageAlreadyExists figures out if an image is already in the local repository.
	Returns true if it is available in the local registry, false otherwise.*/
	CheckIfImageAlreadyExists(dockerImage string, tag string) (bool, error)

	// PullImageFromRegistry pulls an image from the docker registry given a docker Image name, image tag and image architecture
	PullImageFromRegistry(dockerImage string, tag string, arch string) error

	/* CreateContainer creates a a container given a container name, image name, image tag and list of commands for cmd.
	It returns the ID of the new created container */
	CreateContainer(containerName string, image string, tag string, cmd []string) (string, error)

	// RunContainer starts a new container given a container ID.
	RunContainer(containerID string) error

	/* CheckIfContainerIsReady checks if a container is in running state
	It returns true if it is running, false if in any other  */
	CheckIfContainerIsReady(containerID string) (bool, error)

	/* GenerateExecInstance generates a new exec instance on a container given a container ID and a command to run
	It returns the exec ID */
	GenerateExecInstance(containerID string, commands []string) (string, error)

	/* StartExecInstance start an exec instance on a docker daemon given an exec ID.
	It returns the stdout and stderr from inside the container */
	StartExecInstance(execInstanceID string) (string, error)

	/* StopContainer stops a container given a container ID.
	Returns true if the container is stopped and false is the container was already stopped */
	StopContainer(containerID string) (bool, error)

	// RemoveContainer removes a container given a container ID
	RemoveContainer(containerID string) error
}
