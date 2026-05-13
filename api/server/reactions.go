package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireServerReaction is one reaction entry returned by the Server reactions
// endpoint. Each entry represents one (emoji, user) pair.
type wireServerReaction struct {
	Emoticon struct {
		Value string `json:"value"` // e.g. ":thumbsup:" or "thumbsup"
	} `json:"emoticon"`
	User struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
}

// wireServerReactionsPage is the paged envelope for the reactions endpoint.
// Server returns this as a standard paged response.
type wireServerReactionsPage struct {
	Values     []wireServerReaction `json:"values"`
	IsLastPage bool                 `json:"isLastPage"`
	Size       int                  `json:"size"`
}

// ListCommentReactions lists the reactions on a pull-request comment, grouped
// by emoji.
//
// API: GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{prID}/comments/{commentID}/reactions
func (c *Client) ListCommentReactions(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d/reactions", ns, slug, prID, commentID)

	// Collect all pages. Reactions are typically few so we fetch only the first
	// page here; a full paginated loop can be added later if needed.
	var allEntries []wireServerReaction
	var page wireServerReactionsPage
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	allEntries = page.Values

	// Group by canonical emoji.
	order := make([]string, 0)
	groups := make(map[string][]backend.User)
	for _, e := range allEntries {
		emoji := backend.NormaliseEmoji(e.Emoticon.Value)
		if _, exists := groups[emoji]; !exists {
			order = append(order, emoji)
		}
		groups[emoji] = append(groups[emoji], backend.User{
			Slug:        e.User.Slug,
			DisplayName: e.User.DisplayName,
		})
	}

	reactions := make([]backend.CommentReaction, 0, len(order))
	for _, emoji := range order {
		reactions = append(reactions, backend.CommentReaction{
			Emoji: emoji,
			Users: groups[emoji],
		})
	}
	return reactions, nil
}

// wireServerAddReaction is the POST body for adding a reaction.
type wireServerAddReaction struct {
	Emoticon struct {
		Value string `json:"value"`
	} `json:"emoticon"`
}

// AddCommentReaction adds an emoji reaction to a pull-request comment.
//
// API: POST /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{prID}/comments/{commentID}/reactions
func (c *Client) AddCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d/reactions", ns, slug, prID, commentID)
	var body wireServerAddReaction
	body.Emoticon.Value = emoji
	var resp json.RawMessage
	return c.postJSON(path, body, &resp)
}

// RemoveCommentReaction removes an emoji reaction from a pull-request comment.
//
// API: DELETE /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{prID}/comments/{commentID}/reactions/{emoji}
func (c *Client) RemoveCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d/reactions/%s", ns, slug, prID, commentID, emoji)
	return c.delete(path, nil)
}
