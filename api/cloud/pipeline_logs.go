package cloud

import (
	"fmt"
	"io"
)

// GetPipelineStepLog returns a streaming reader for the plaintext log of a
// single pipeline step. The caller is responsible for calling Close on the
// returned reader.
func (c *Client) GetPipelineStepLog(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error) {
	path := fmt.Sprintf(
		"/repositories/%s/%s/pipelines/%s/steps/%s/log",
		ns, slug, braceUUID(pipelineUUID), braceUUID(stepUUID),
	)
	return c.http.GetStream(path)
}
