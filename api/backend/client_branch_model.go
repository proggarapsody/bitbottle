package backend

// BranchModelClient is implemented only by Bitbucket Cloud clients.
// Bitbucket Server / Data Center has no branch model API.
type BranchModelClient interface {
	GetBranchModel(ws, slug string) (BranchModel, error)
	GetBranchModelSettings(ws, slug string) (BranchModelSettings, error)
	UpdateBranchModelSettings(ws, slug string, in BranchModelSettingsInput) (BranchModelSettings, error)
}

// FeatureBranchModel names the branch-model capability for typed-error reporting.
const FeatureBranchModel Feature = "branch_model"

// AsBranchModelClient returns the BranchModelClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsBranchModelClient(c Client, host string) (BranchModelClient, error) {
	return requireFeature[BranchModelClient](c, host, specFor(FeatureBranchModel))
}

// BranchModel is the effective branching model for a repository.
type BranchModel struct {
	Development BranchModelBranch  `json:"development"`
	Production  *BranchModelBranch `json:"production,omitempty"`
	BranchTypes []BranchType       `json:"branch_types"`
}

// BranchModelBranch describes a development or production branch in the model.
type BranchModelBranch struct {
	Name          string `json:"name"`
	IsValid       bool   `json:"is_valid"`
	UseMainbranch bool   `json:"use_mainbranch"`
}

// BranchType is a named branch type with a prefix used by the "Create branch" wizard.
type BranchType struct {
	Kind   string `json:"kind"`
	Prefix string `json:"prefix"`
}

// BranchModelSettings is the editable branching model configuration.
type BranchModelSettings struct {
	Development BranchModelSettingsBranch `json:"development"`
	Production  BranchModelSettingsBranch `json:"production"`
	BranchTypes []BranchTypeSettings      `json:"branch_types"`
}

// BranchModelSettingsBranch describes a configured branch in the settings.
type BranchModelSettingsBranch struct {
	IsValid       bool   `json:"is_valid"`
	Name          string `json:"name"`
	UseMainbranch bool   `json:"use_mainbranch"`
}

// BranchTypeSettings is the per-kind configuration for the "Create branch" wizard.
type BranchTypeSettings struct {
	Enabled bool   `json:"enabled"`
	Kind    string `json:"kind"`
	Prefix  string `json:"prefix"`
}

// BranchModelSettingsInput is the PUT body for updating branching model settings.
type BranchModelSettingsInput struct {
	Development *BranchModelSettingsBranch `json:"development,omitempty"`
	Production  *BranchModelSettingsBranch `json:"production,omitempty"`
	BranchTypes []BranchTypeSettings       `json:"branch_types,omitempty"`
}
