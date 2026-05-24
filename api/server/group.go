package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// restGroup is the wire representation of a Bitbucket Server/DC group.
type restGroup struct {
	Name string `json:"name"`
}

// restGroupMember is the wire representation of a group member returned by
// the more-members endpoint.
type restGroupMember struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

func toGroupDomain(w restGroup) backend.Group {
	return backend.Group{Name: w.Name}
}

func toGroupMemberDomain(w restGroupMember) backend.GroupMember {
	return backend.GroupMember{
		Name:         w.Name,
		DisplayName:  w.DisplayName,
		EmailAddress: w.EmailAddress,
	}
}

// ListGroups returns admin groups on Bitbucket Server/DC, optionally
// filtered by a name prefix.
func (c *Client) ListGroups(filter string, limit int) ([]backend.Group, error) {
	path := fmt.Sprintf("/admin/groups?limit=%d", limit)
	if filter != "" {
		path += "&filter=" + url.QueryEscape(filter)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Group, error) {
		var page PagedResponse[restGroup]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Group, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toGroupDomain(w))
		}
		return out, nil
	}, limit)
}

// CreateGroup creates an admin group with the given name.
func (c *Client) CreateGroup(name string) (backend.Group, error) {
	type createRequest struct {
		Name string `json:"name"`
	}
	var wire restGroup
	if err := c.postJSON("/admin/groups", createRequest{Name: name}, &wire); err != nil {
		return backend.Group{}, err
	}
	return toGroupDomain(wire), nil
}

// DeleteGroup deletes the admin group with the given name.
func (c *Client) DeleteGroup(name string) error {
	path := "/admin/groups?name=" + url.QueryEscape(name)
	return c.delete(path, nil)
}

// ListGroupMembers returns the members of the named admin group.
func (c *Client) ListGroupMembers(groupName string, limit int) ([]backend.GroupMember, error) {
	path := fmt.Sprintf("/admin/groups/more-members?context=%s&limit=%d",
		url.QueryEscape(groupName), limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.GroupMember, error) {
		var page PagedResponse[restGroupMember]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.GroupMember, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toGroupMemberDomain(w))
		}
		return out, nil
	}, limit)
}

// AddGroupMember adds user to the named admin group.
func (c *Client) AddGroupMember(groupName, user string) error {
	type addRequest struct {
		User  string `json:"user"`
		Group string `json:"group"`
	}
	return c.postJSON("/admin/users/add-group", addRequest{User: user, Group: groupName}, nil)
}

// RemoveGroupMember removes user from the named admin group.
func (c *Client) RemoveGroupMember(groupName, user string) error {
	type removeRequest struct {
		User  string `json:"user"`
		Group string `json:"group"`
	}
	return c.postJSON("/admin/users/remove-group", removeRequest{User: user, Group: groupName}, nil)
}
