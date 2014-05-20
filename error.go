package unhtml

type ErrUnmarshal struct {
	Reason, Info string
}

func NewErr(r, i string) error {
	return &ErrUnmarshal{r, i}
}

func (e *ErrUnmarshal) Error() string {
	s := "Error unmarshaling html (" + e.Reason + ")"
	if e.Info != "" {
		s += ": " + e.Info
	}
	return s
}
