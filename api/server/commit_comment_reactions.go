package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ListCommitCommentReactions lists the reactions on a commit comment, grouped
// by emoji.
//
// API: GET /rest/api/1.0/projects/{ns}/repos/{slug}/commits/{hash}/comments/{commentID}/reactions
func (c *Client) ListCommitCommentReactions(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d/reactions", ns, slug, hash, commentID)

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

// AddCommitCommentReaction adds an emoji reaction to a commit comment.
//
// API: POST /rest/api/1.0/projects/{ns}/repos/{slug}/commits/{hash}/comments/{commentID}/reactions
func (c *Client) AddCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d/reactions", ns, slug, hash, commentID)
	var body wireServerAddReaction
	body.Emoticon.Value = emoji
	var resp json.RawMessage
	return c.postJSON(path, body, &resp)
}

// RemoveCommitCommentReaction removes an emoji reaction from a commit comment.
//
// API: DELETE /rest/api/1.0/projects/{ns}/repos/{slug}/commits/{hash}/comments/{commentID}/reactions/{emoji}
func (c *Client) RemoveCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d/reactions/%s", ns, slug, hash, commentID, emoji)
	return c.delete(path, nil)
}
