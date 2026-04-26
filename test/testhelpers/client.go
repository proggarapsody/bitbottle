package testhelpers

import (
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// FakeClient is a test double for backend.Client.
// Set the Fn fields for the methods your test exercises.
// Unset methods call t.Fatalf so unexpected calls are loud failures.
type FakeClient struct {
	T *testing.T

	// Repo methods
	ListReposFn  func(limit int) ([]backend.Repository, error)
	GetRepoFn    func(ns, slug string) (backend.Repository, error)
	CreateRepoFn func(ns string, in backend.CreateRepoInput) (backend.Repository, error)
	DeleteRepoFn func(ns, slug string) error

	// PR methods
	ListPRsFn   func(ns, slug, state string, limit int) ([]backend.PullRequest, error)
	GetPRFn     func(ns, slug string, id int) (backend.PullRequest, error)
	CreatePRFn  func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error)
	MergePRFn   func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error)
	ApprovePRFn func(ns, slug string, id int) error
	GetPRDiffFn func(ns, slug string, id int) (string, error)

	// Branch / user methods
	ListBranchesFn   func(ns, slug string, limit int) ([]backend.Branch, error)
	DeleteBranchFn   func(ns, slug, branch string) error
	GetCurrentUserFn func() (backend.User, error)

	// Branch create method
	CreateBranchFn func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error)

	// Tag methods
	ListTagsFn   func(ns, slug string, limit int) ([]backend.Tag, error)
	CreateTagFn  func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error)
	DeleteTagFn  func(ns, slug, name string) error

	// PR lifecycle methods
	UpdatePRFn        func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error)
	DeclinePRFn       func(ns, slug string, id int) error
	UnapprovePRFn     func(ns, slug string, id int) error
	ReadyPRFn         func(ns, slug string, id int) error
	RequestReviewFn   func(ns, slug string, id int, users []string) error

	// Pipeline methods (Cloud-only; satisfies backend.PipelineClient when set)
	ListPipelinesFn func(ns, slug string, limit int) ([]backend.Pipeline, error)
	GetPipelineFn   func(ns, slug, uuid string) (backend.Pipeline, error)
	RunPipelineFn   func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error)

	// Commit methods
	ListCommitsFn func(ns, slug, branch string, limit int) ([]backend.Commit, error)
	GetCommitFn   func(ns, slug, hash string) (backend.Commit, error)
}

// Compile-time interface check.
var _ backend.Client = (*FakeClient)(nil)

func (c *FakeClient) ListRepos(limit int) ([]backend.Repository, error) {
	if c.ListReposFn != nil {
		return c.ListReposFn(limit)
	}
	return nil, nil
}

func (c *FakeClient) GetRepo(ns, slug string) (backend.Repository, error) {
	if c.GetRepoFn != nil {
		return c.GetRepoFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetRepo; set GetRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	if c.CreateRepoFn != nil {
		return c.CreateRepoFn(ns, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateRepo; set CreateRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) DeleteRepo(ns, slug string) error {
	if c.DeleteRepoFn != nil {
		return c.DeleteRepoFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteRepo; set DeleteRepoFn in your test")
	}
	return nil
}

func (c *FakeClient) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	if c.ListPRsFn != nil {
		return c.ListPRsFn(ns, slug, state, limit)
	}
	return nil, nil
}

func (c *FakeClient) GetPR(ns, slug string, id int) (backend.PullRequest, error) {
	if c.GetPRFn != nil {
		return c.GetPRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPR; set GetPRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	if c.CreatePRFn != nil {
		return c.CreatePRFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreatePR; set CreatePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	if c.MergePRFn != nil {
		return c.MergePRFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.MergePR; set MergePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) ApprovePR(ns, slug string, id int) error {
	if c.ApprovePRFn != nil {
		return c.ApprovePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ApprovePR; set ApprovePRFn in your test")
	}
	return nil
}

func (c *FakeClient) GetPRDiff(ns, slug string, id int) (string, error) {
	if c.GetPRDiffFn != nil {
		return c.GetPRDiffFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPRDiff; set GetPRDiffFn in your test")
	}
	return "", nil
}

func (c *FakeClient) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	if c.ListBranchesFn != nil {
		return c.ListBranchesFn(ns, slug, limit)
	}
	return nil, nil
}

func (c *FakeClient) DeleteBranch(ns, slug, branch string) error {
	if c.DeleteBranchFn != nil {
		return c.DeleteBranchFn(ns, slug, branch)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteBranch; set DeleteBranchFn in your test")
	}
	return nil
}

func (c *FakeClient) GetCurrentUser() (backend.User, error) {
	if c.GetCurrentUserFn != nil {
		return c.GetCurrentUserFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetCurrentUser; set GetCurrentUserFn in your test")
	}
	return backend.User{}, nil
}

func (c *FakeClient) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	if c.CreateBranchFn != nil {
		return c.CreateBranchFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateBranch; set CreateBranchFn in your test")
	}
	return backend.Branch{}, nil
}

func (c *FakeClient) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	if c.ListTagsFn != nil {
		return c.ListTagsFn(ns, slug, limit)
	}
	return nil, nil
}

func (c *FakeClient) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	if c.CreateTagFn != nil {
		return c.CreateTagFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateTag; set CreateTagFn in your test")
	}
	return backend.Tag{}, nil
}

func (c *FakeClient) DeleteTag(ns, slug, name string) error {
	if c.DeleteTagFn != nil {
		return c.DeleteTagFn(ns, slug, name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteTag; set DeleteTagFn in your test")
	}
	return nil
}

func (c *FakeClient) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	if c.UpdatePRFn != nil {
		return c.UpdatePRFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdatePR; set UpdatePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) DeclinePR(ns, slug string, id int) error {
	if c.DeclinePRFn != nil {
		return c.DeclinePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeclinePR; set DeclinePRFn in your test")
	}
	return nil
}

func (c *FakeClient) UnapprovePR(ns, slug string, id int) error {
	if c.UnapprovePRFn != nil {
		return c.UnapprovePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UnapprovePR; set UnapprovePRFn in your test")
	}
	return nil
}

func (c *FakeClient) ReadyPR(ns, slug string, id int) error {
	if c.ReadyPRFn != nil {
		return c.ReadyPRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ReadyPR; set ReadyPRFn in your test")
	}
	return nil
}

func (c *FakeClient) RequestReview(ns, slug string, id int, users []string) error {
	if c.RequestReviewFn != nil {
		return c.RequestReviewFn(ns, slug, id, users)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RequestReview; set RequestReviewFn in your test")
	}
	return nil
}

func (c *FakeClient) ListPipelines(ns, slug string, limit int) ([]backend.Pipeline, error) {
	if c.ListPipelinesFn != nil {
		return c.ListPipelinesFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelines; set ListPipelinesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetPipeline(ns, slug, uuid string) (backend.Pipeline, error) {
	if c.GetPipelineFn != nil {
		return c.GetPipelineFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipeline; set GetPipelineFn in your test")
	}
	return backend.Pipeline{}, nil
}

func (c *FakeClient) RunPipeline(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
	if c.RunPipelineFn != nil {
		return c.RunPipelineFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RunPipeline; set RunPipelineFn in your test")
	}
	return backend.Pipeline{}, nil
}

func (c *FakeClient) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	if c.ListCommitsFn != nil {
		return c.ListCommitsFn(ns, slug, branch, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListCommits; set ListCommitsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	if c.GetCommitFn != nil {
		return c.GetCommitFn(ns, slug, hash)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetCommit; set GetCommitFn in your test")
	}
	return backend.Commit{}, nil
}
