package models

// CreateContainerBody is the struct that models request body when creating a container
type CreateContainerBody struct {
	// Cmd is the list of commands to execute when creating the container
	Cmd []string
	// Image the the image plus tag to use
	Image string
}

type GenerateExecInstanceBody struct {
	// AttachStdin attaches to stdin
	AttachStdin bool
	// AttachStdout attaches to stdout
	AttachStdout bool
	// AttachStderr attaches to stderr
	AttachStderr bool
	// Tty Attach standard streams to a tty, including stdin if it is not closed
	Tty bool
	// Cmd are the commands to run inside the container
	Cmd []string
}

type StartExecInstance struct {
	// Detach mode: run command in the background
	Detach bool
	// Tty Allocate a pseudo-TTY
	Tty bool
}
