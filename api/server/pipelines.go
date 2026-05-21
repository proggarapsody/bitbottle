package server

import "github.com/proggarapsody/bitbottle/api/backend"

// StopPipeline always returns ErrUnsupportedOnHost because Bitbucket Server
// does not have a Pipelines API.
func (c *Client) StopPipeline(_, _, _ string) error {
	return &backend.DomainError{Kind: backend.ErrUnsupportedOnHost, Message: "pipeline stop is not supported on Bitbucket Server"}
}
