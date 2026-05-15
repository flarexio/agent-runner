package runner

import "errors"

const (
	EventIssue       = "issue"
	EventIssueClosed = "issue-closed"

	IssueMarker = "<!-- agent-task:v1 -->"

	LabelTypeAgentTask     = "type:agent-task"
	LabelAgentClaudeCode   = "agent:claude-code"
	LabelAgentReady        = "agent-ready"
	LabelAgentApproved     = "agent-approved"
	LabelClaimedByClaude   = "claimed-by-claude"
	LabelAgentBlocked      = "agent-blocked"
	LabelSecuritySensitive = "security-sensitive"
	LabelAgentFailed       = "agent-failed"

	LabelModelFast     = "model:fast"
	LabelModelBalanced = "model:balanced"
	LabelModelStrong   = "model:strong"
)

var (
	RequiredIssueLabels = []string{
		LabelTypeAgentTask,
		LabelAgentClaudeCode,
		LabelAgentReady,
		LabelAgentApproved,
	}

	ExcludedIssueLabels = []string{
		LabelClaimedByClaude,
		LabelAgentBlocked,
		LabelSecuritySensitive,
	}

	KnownModelLabels = []string{
		LabelModelFast,
		LabelModelBalanced,
		LabelModelStrong,
	}
)

var (
	ErrTaskNotFound             = errors.New("task not found")
	ErrInvalidPrompt            = errors.New("invalid prompt")
	ErrExecFailed               = errors.New("claude execution failed")
	ErrInvalidRepo              = errors.New("invalid repo")
	ErrInvalidIssueNumber       = errors.New("invalid issue number")
	ErrIssueNotOpen             = errors.New("issue is not open")
	ErrIssueMarkerMissing       = errors.New("issue body missing agent task marker")
	ErrIssueLabelMissing        = errors.New("issue missing required label")
	ErrIssueLabelExcluded       = errors.New("issue has excluded label")
	ErrIssueMultipleModelLabels = errors.New("issue has multiple model recommendation labels")
	ErrGitHubUnavailable        = errors.New("github client not configured")
	ErrIssueWorkspaceBusy       = errors.New("issue workspace already locked")
)

type Config struct {
	WorkDir      string       `yaml:"workDir"`
	AllowedTools []string     `yaml:"allowedTools"`
	MaxTurns     int          `yaml:"maxTurns"`
	Model        string       `yaml:"model,omitempty"`
	Issue        EventConfig  `yaml:"issue,omitempty"`
	GitHub       GitHubConfig `yaml:"github,omitempty"`
}

// EventConfig overrides top-level Claude flags for an event. Empty fields fall
// back to the top-level Config values.
//
// BypassPermissions passes --dangerously-skip-permissions and ignores
// AllowedTools — only for trusted unattended flows.
type EventConfig struct {
	AllowedTools      []string          `yaml:"allowedTools,omitempty"`
	MaxTurns          int               `yaml:"maxTurns,omitempty"`
	Model             string            `yaml:"model,omitempty"`
	ModelLabels       map[string]string `yaml:"modelLabels,omitempty"`
	BypassPermissions bool              `yaml:"bypassPermissions,omitempty"`
}

type GitHubConfig struct {
	Token   string `yaml:"token,omitempty"`
	BaseURL string `yaml:"baseURL,omitempty"`
}
