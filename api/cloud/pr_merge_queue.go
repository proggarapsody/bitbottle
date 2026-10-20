package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func (c *Client) queueError(code backend.ErrorCode, id int, message string) error {
	return &backend.DomainError{Kind: backend.ErrConflict, Code: code, Host: hostFromURL(c.baseURL), Resource: "pull-request", ID: strconv.Itoa(id), Message: message}
}

func queueRepoPath(ns, slug string) string {
	return fmt.Sprintf("/repositories/%s/%s", url.PathEscape(ns), url.PathEscape(slug))
}

func (c *Client) mergeQueues(ns, slug, branch string) ([]cloudgen.CloudMergeQueue, error) {
	path := queueRepoPath(ns, slug) + "/merge-queues?target_branch=" + url.QueryEscape(branch) + "&pagelen=100"
	return paging.Collect(c.http, path, func(body []byte) ([]cloudgen.CloudMergeQueue, error) {
		var page cloudPagedResponse[cloudgen.CloudMergeQueue]
		err := json.Unmarshal(body, &page)
		return page.Values, err
	}, 0)
}

func queueItemPath(ns, slug, uuid string, id int) string {
	return fmt.Sprintf("%s/merge-queues/%s/queue/%d", queueRepoPath(ns, slug), url.PathEscape(uuid), id)
}

func (c *Client) queueEntry(ns, slug, uuid string, id int) (*cloudgen.CloudQueuedPullRequest, error) {
	var item cloudgen.CloudQueuedPullRequest
	err := c.getJSON(queueItemPath(ns, slug, uuid, id), &item)
	var de *backend.DomainError
	if errors.As(err, &de) && de.HTTPStatus() == 404 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if item.Pullrequest == nil || item.Pullrequest.ID != id ||
		(item.MergeQueue != nil && item.MergeQueue.UUID != nil && *item.MergeQueue.UUID != uuid) {
		return nil, c.queueError(backend.CodePRMergeQueueUnconfirmed, id, "merge queue returned an unexpected pull request or queue identity")
	}
	return &item, nil
}

func (c *Client) queueContains(ns, slug, uuid string, id int) (bool, error) {
	item, err := c.queueEntry(ns, slug, uuid, id)
	return item != nil, err
}

// EnableAutoMerge requires a destination queue before submitting the ordinary
// merge operation: without that guard the same endpoint can merge immediately.
func (c *Client) EnableAutoMerge(ns, slug string, id int, strategy string) error {
	pr, err := c.GetPR(ns, slug, id)
	if err != nil {
		return err
	}
	if pr.State != "OPEN" {
		return c.queueError(backend.CodePRTerminalState, id, "automatic merge requires an open pull request")
	}
	queues, err := c.mergeQueues(ns, slug, pr.ToBranch)
	if err != nil {
		return err
	}
	var uuid string
	for _, queue := range queues {
		if queue.State == "ACTIVE" {
			uuid = queue.UUID
			break
		}
	}
	if uuid == "" {
		return c.queueError(backend.CodePRMergeQueueRequired, id, "no active merge queue is configured for destination branch "+pr.ToBranch)
	}
	mapped := strategy
	switch strategy {
	case "", "merge", "merge-commit":
		mapped = "merge_commit"
	case "rebase":
		mapped = "rebase_merge"
	}
	if err := c.postJSON(fmt.Sprintf("%s/pullrequests/%d/merge", queueRepoPath(ns, slug), id), cloudgen.CloudMergePullRequest{MergeStrategy: &mapped}, nil); err != nil {
		return stampPRMerge(err, id)
	}
	for attempt, deadline := 0, time.Now().Add(30*time.Second); time.Now().Before(deadline); attempt++ {
		if attempt > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		item, err := c.queueEntry(ns, slug, uuid, id)
		if err != nil {
			return err
		}
		if item != nil {
			if item.Status != nil {
				switch *item.Status {
				case "QUEUED", "PROCESSING", "AWAITING_MERGE", "MERGE_IN_PROGRESS":
					return nil
				}
			}
			return c.queueError(backend.CodePRMergeQueueUnconfirmed, id, "merge queue entry is being removed or has an unconfirmed lifecycle state")
		}
		current, err := c.GetPR(ns, slug, id)
		if err != nil {
			return err
		}
		if current.State == "MERGED" {
			return nil
		}
		if current.State != "OPEN" {
			return c.queueError(backend.CodePRTerminalState, id, "pull request closed before queueing could be confirmed")
		}
	}
	return c.queueError(backend.CodePRMergeQueueUnconfirmed, id, "merge request accepted but queue membership was not confirmed")
}

// DisableAutoMerge searches every queue state because a draining or suspended
// queue can still hold an entry that must be cancelled.
func (c *Client) DisableAutoMerge(ns, slug string, id int) error {
	pr, err := c.GetPR(ns, slug, id)
	if err != nil {
		return err
	}
	if pr.State != "OPEN" {
		return c.queueError(backend.CodePRTerminalState, id, "cancellation requires an open pull request; a completed merge cannot be cancelled")
	}
	queues, err := c.mergeQueues(ns, slug, pr.ToBranch)
	if err != nil {
		return err
	}
	if len(queues) == 0 {
		return c.queueError(backend.CodePRMergeQueueRequired, id, "no merge queue is configured for destination branch "+pr.ToBranch)
	}
	for _, queue := range queues {
		present, err := c.queueContains(ns, slug, queue.UUID, id)
		if err != nil {
			return err
		}
		if !present {
			continue
		}
		if err := c.delete(queueItemPath(ns, slug, queue.UUID, id)); err != nil {
			return err
		}
		removed := false
		for attempt, deadline := 0, time.Now().Add(30*time.Second); time.Now().Before(deadline); attempt++ {
			if attempt > 0 {
				time.Sleep(200 * time.Millisecond)
			}
			present, err = c.queueContains(ns, slug, queue.UUID, id)
			if err != nil {
				return err
			}
			if !present {
				removed = true
				break
			}
		}
		if !removed {
			return c.queueError(backend.CodePRMergeQueueUnconfirmed, id, "cancellation accepted but the pull request remains in the merge queue")
		}
	}
	current, err := c.GetPR(ns, slug, id)
	if err != nil {
		return err
	}
	if current.State != "OPEN" {
		return c.queueError(backend.CodePRTerminalState, id, "the pull request closed while cancellation was in progress; the merge may have completed")
	}
	if current.AutoMerge != nil && current.AutoMerge.Enabled {
		return c.queueError(backend.CodePRMergeQueueUnconfirmed, id, "the pull request still reports queued after cancellation")
	}
	return nil
}
