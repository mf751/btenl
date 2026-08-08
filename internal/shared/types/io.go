package types

// INFO: Types used by the controls layer and daemon to communicate
type ControlRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ControlResponseStatus uint8

const (
	StatusSucceed ControlResponseStatus = iota
	StatusDaemonDied
	StatusDaemonError
	StatusInvalidRequest
	StatusTimout
)

type ControlResponse struct {
	Status ControlResponseStatus `json:"status"`
	Msg    string                `json:"message"`
}
