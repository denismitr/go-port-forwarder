package portforwarder

import "errors"

var (
	ErrTargetPodValidation = errors.New("target pod validation failed")
	ErrPodNotFound         = errors.New("could not find pod to forward ports")
)
