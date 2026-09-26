package organism

// Status is the observable lifecycle state of an organism.
type Status string

const (
	StatusCreated    Status = "created"
	StatusReady      Status = "ready"
	StatusRunning    Status = "running"
	StatusStopped    Status = "stopped"
	StatusError      Status = "error"
	StatusRecovering Status = "recovering"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusCreated, StatusReady, StatusRunning, StatusStopped, StatusError, StatusRecovering:
		return true
	default:
		return false
	}
}
