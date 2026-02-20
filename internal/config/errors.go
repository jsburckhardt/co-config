package config

import "errors"

var (
	ErrConfigNotFound = errors.New("copilot config file not found")
	ErrConfigInvalid  = errors.New("copilot config file is invalid")
)
