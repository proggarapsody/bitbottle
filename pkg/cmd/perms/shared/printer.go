// Package shared provides shared helpers for the perms command tree.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// GrantPrinter returns a Printer configured for PermissionGrant display.
func GrantPrinter(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PermissionGrant] {
	p := format.New[backend.PermissionGrant](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.PermissionGrant]{
		Name:   "permission",
		Header: "PERMISSION",
		Extract: func(g backend.PermissionGrant) any {
			return g.Permission
		},
	})
	p.AddField(format.Field[backend.PermissionGrant]{
		Name:   "kind",
		Header: "KIND",
		Extract: func(g backend.PermissionGrant) any {
			return g.Subject.Kind
		},
	})
	p.AddField(format.Field[backend.PermissionGrant]{
		Name:   "name",
		Header: "NAME",
		Extract: func(g backend.PermissionGrant) any {
			if g.Subject.Kind == "user" {
				return g.Subject.Slug
			}
			return g.Subject.Name
		},
	})
	p.AddField(format.Field[backend.PermissionGrant]{
		Name:     "display_name",
		Header:   "DISPLAY NAME",
		JSONOnly: false,
		Extract: func(g backend.PermissionGrant) any {
			if g.Subject.DisplayName == "" {
				return nil
			}
			return g.Subject.DisplayName
		},
	})
	return p
}

// SubjectFromFlags constructs a PermissionSubject from the --user / --group
// flag values. Exactly one of userSlug and groupName must be non-empty.
func SubjectFromFlags(userSlug, groupName string) (backend.PermissionSubject, error) {
	if (userSlug == "") == (groupName == "") {
		return backend.PermissionSubject{}, &subjectFlagError{}
	}
	if userSlug != "" {
		return backend.PermissionSubject{Kind: "user", Slug: userSlug}, nil
	}
	return backend.PermissionSubject{Kind: "group", Name: groupName}, nil
}

// subjectFlagError is returned when neither or both --user / --group are set.
type subjectFlagError struct{}

func (*subjectFlagError) Error() string {
	return "specify exactly one of --user or --group (e.g. --user alice or --group devs)"
}
