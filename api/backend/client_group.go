package backend

// GroupClient manages Bitbucket Server/DC admin groups. This capability is
// Server-only; Cloud returns ErrUnsupportedOnHost for all methods.
type GroupClient interface {
	ListGroups(filter string, limit int) ([]Group, error)
	CreateGroup(name string) (Group, error)
	DeleteGroup(name string) error
}

// GroupMemberClient manages members of Bitbucket Server/DC admin groups.
// This capability is Server-only; Cloud returns ErrUnsupportedOnHost.
type GroupMemberClient interface {
	ListGroupMembers(groupName string, limit int) ([]GroupMember, error)
	AddGroupMember(groupName, user string) error
	RemoveGroupMember(groupName, user string) error
}

// Group is the domain representation of a Bitbucket Server/DC admin group.
type Group struct {
	Name string `json:"name"`
}

// GroupMember is a user that belongs to a Bitbucket Server/DC admin group.
type GroupMember struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// FeatureGroup names the group management capability for typed-error reporting.
const FeatureGroup Feature = "group"

// FeatureGroupMember names the group-member management capability.
const FeatureGroupMember Feature = "group_member"

// AsGroupClient returns the GroupClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the Group capability.
func AsGroupClient(c Client, host string) (GroupClient, error) {
	return requireFeature[GroupClient](c, host, specFor(FeatureGroup))
}

// AsGroupMemberClient returns the GroupMemberClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the GroupMember capability.
func AsGroupMemberClient(c Client, host string) (GroupMemberClient, error) {
	return requireFeature[GroupMemberClient](c, host, specFor(FeatureGroupMember))
}
