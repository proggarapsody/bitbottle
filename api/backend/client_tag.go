package backend

// TagLister lists tags in a repository.
type TagLister interface {
	ListTags(ns, slug string, limit int) ([]Tag, error)
}

// TagCreator creates a tag.
type TagCreator interface {
	CreateTag(ns, slug string, in CreateTagInput) (Tag, error)
}

// TagDeleter deletes a tag.
type TagDeleter interface {
	DeleteTag(ns, slug, name string) error
}
