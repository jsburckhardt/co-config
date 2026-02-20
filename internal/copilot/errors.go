package copilot

import "errors"

var (
	ErrCopilotNotInstalled = errors.New("copilot CLI not installed")
	ErrVersionParseFailed  = errors.New("failed to parse copilot version")
	ErrSchemaParseFailed   = errors.New("failed to parse copilot config schema")
)
