package models

/* CreateContainerResponseBody wraps the response body coming from the docker daemon when
   creating a container */
type CreateContainerResponseBody struct {
	// ID of the created container
	ID string
	// Warnings is a list of warnings that may occur
	Warnings []string
}

type CheckContainerStatusBody struct {
	// State is an object that gives us different info about the container
	State struct {
		// Status gives us the current container status (running, exited, paused)
		Status string
		// Running tells if the container is running or not
		Running bool
	}
}

type CreateExecResponseBody struct {
	// ID of the created exec instance
	ID string
}
