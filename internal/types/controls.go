package types

type ControlRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ControlResponseState uint8

const (
	StatusSucceed = iota
	StatusDaemonDied
	StatusDaemonError
	StatusInvalidRequest
	StatusTimout
)

type ControlResponse struct {
	Status ControlResponseState `json:"status"`
	Msg    string               `json:"message"`
}

type ControlEvent struct {
	Command string
	Args    []string

	Reply func(ControlResponse) error
}

func (ControlEvent) isEvent() {}
