package github

// ClientInterface defines the interface for a GitHub client
type ClientInterface interface {
	// GetRepository gets information about a repository
	GetRepository(owner, name string) (*Repository, error)

	// ListPullRequests lists pull requests for a repository
	ListPullRequests(owner, name string, options *PullRequestOptions) ([]*PullRequest, error)

	// ListIssues lists issues for a repository
	ListIssues(owner, name string, options *IssueOptions) ([]*Issue, error)

	// GetRateLimit gets the current GitHub API rate limit
	GetRateLimit() (*RateLimit, error)
}
