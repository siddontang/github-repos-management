package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/siddontang/github-repos-management/internal/models"
)

// Cache is an in-memory implementation of the cache.Cache interface
type Cache struct {
	mu           sync.RWMutex
	repositories map[string]*models.Repository          // key: fullName (owner/name)
	pullRequests map[string]map[int]*models.PullRequest // key: repoFullName -> number
	issues       map[string]map[int]*models.Issue       // key: repoFullName -> number
	labels       map[string]*models.Label               // key: name
	prLabels     map[string]map[int]map[string]struct{} // key: repoFullName -> prNumber -> labelName
	issueLabels  map[string]map[int]map[string]struct{} // key: repoFullName -> issueNumber -> labelName
}

// NewCache creates a new in-memory cache
func NewCache() *Cache {
	return &Cache{
		repositories: make(map[string]*models.Repository),
		pullRequests: make(map[string]map[int]*models.PullRequest),
		issues:       make(map[string]map[int]*models.Issue),
		labels:       make(map[string]*models.Label),
		prLabels:     make(map[string]map[int]map[string]struct{}),
		issueLabels:  make(map[string]map[int]map[string]struct{}),
	}
}

// Repository operations

// AddRepository adds a repository to the cache
func (c *Cache) AddRepository(ctx context.Context, repo *models.Repository) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository already exists
	fullName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	if _, exists := c.repositories[fullName]; exists {
		return fmt.Errorf("repository %s already exists", fullName)
	}

	// Add repository
	c.repositories[fullName] = repo

	// Initialize maps for pull requests and issues
	if _, exists := c.pullRequests[fullName]; !exists {
		c.pullRequests[fullName] = make(map[int]*models.PullRequest)
	}
	if _, exists := c.issues[fullName]; !exists {
		c.issues[fullName] = make(map[int]*models.Issue)
	}
	if _, exists := c.prLabels[fullName]; !exists {
		c.prLabels[fullName] = make(map[int]map[string]struct{})
	}
	if _, exists := c.issueLabels[fullName]; !exists {
		c.issueLabels[fullName] = make(map[int]map[string]struct{})
	}

	return nil
}

// GetRepository gets a repository from the cache
func (c *Cache) GetRepository(ctx context.Context, owner, name string) (*models.Repository, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullName := fmt.Sprintf("%s/%s", owner, name)
	repo, exists := c.repositories[fullName]
	if !exists {
		return nil, fmt.Errorf("repository %s not found", fullName)
	}

	return repo, nil
}

// UpdateRepository updates a repository in the cache
func (c *Cache) UpdateRepository(ctx context.Context, repo *models.Repository) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	fullName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	if _, exists := c.repositories[fullName]; !exists {
		return fmt.Errorf("repository %s not found", fullName)
	}

	// Update repository
	c.repositories[fullName] = repo

	return nil
}

// DeleteRepository deletes a repository from the cache
func (c *Cache) DeleteRepository(ctx context.Context, owner, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullName := fmt.Sprintf("%s/%s", owner, name)
	if _, exists := c.repositories[fullName]; !exists {
		return fmt.Errorf("repository %s not found", fullName)
	}

	// Delete repository
	delete(c.repositories, fullName)

	// Delete associated pull requests
	delete(c.pullRequests, fullName)

	// Delete associated issues
	delete(c.issues, fullName)

	// Delete associated PR labels
	delete(c.prLabels, fullName)

	// Delete associated issue labels
	delete(c.issueLabels, fullName)

	return nil
}

// ListRepositories lists all repositories
func (c *Cache) ListRepositories(ctx context.Context, page, perPage int) ([]*models.Repository, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := len(c.repositories)
	if total == 0 {
		return []*models.Repository{}, 0, nil
	}

	// Convert map to slice
	repos := make([]*models.Repository, 0, total)
	for _, repo := range c.repositories {
		repos = append(repos, repo)
	}

	// Apply pagination
	start := (page - 1) * perPage
	if start >= total {
		return []*models.Repository{}, total, nil
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return repos[start:end], total, nil
}

// Pull request operations

// AddPullRequest adds a pull request to the cache
func (c *Cache) AddPullRequest(ctx context.Context, pr *models.PullRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[pr.RepositoryFullName]; !exists {
		return fmt.Errorf("repository %s not found", pr.RepositoryFullName)
	}

	// Initialize map for repository if it doesn't exist
	if _, exists := c.pullRequests[pr.RepositoryFullName]; !exists {
		c.pullRequests[pr.RepositoryFullName] = make(map[int]*models.PullRequest)
	}

	// Check if pull request already exists
	if _, exists := c.pullRequests[pr.RepositoryFullName][pr.Number]; exists {
		return fmt.Errorf("pull request %s#%d already exists", pr.RepositoryFullName, pr.Number)
	}

	// Add pull request
	c.pullRequests[pr.RepositoryFullName][pr.Number] = pr

	return nil
}

// GetPullRequest gets a pull request from the cache
func (c *Cache) GetPullRequest(ctx context.Context, repoFullName string, number int) (*models.PullRequest, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return nil, fmt.Errorf("no pull requests found for repository %s", repoFullName)
	}

	pr, exists := repoMap[number]
	if !exists {
		return nil, fmt.Errorf("pull request %s#%d not found", repoFullName, number)
	}

	return pr, nil
}

// UpdatePullRequest updates a pull request in the cache
func (c *Cache) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[pr.RepositoryFullName]; !exists {
		return fmt.Errorf("repository %s not found", pr.RepositoryFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[pr.RepositoryFullName]
	if !exists {
		return fmt.Errorf("no pull requests found for repository %s", pr.RepositoryFullName)
	}

	if _, exists := repoMap[pr.Number]; !exists {
		return fmt.Errorf("pull request %s#%d not found", pr.RepositoryFullName, pr.Number)
	}

	// Update pull request
	c.pullRequests[pr.RepositoryFullName][pr.Number] = pr

	return nil
}

// DeletePullRequest deletes a pull request from the cache
func (c *Cache) DeletePullRequest(ctx context.Context, repoFullName string, number int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return fmt.Errorf("no pull requests found for repository %s", repoFullName)
	}

	if _, exists := repoMap[number]; !exists {
		return fmt.Errorf("pull request %s#%d not found", repoFullName, number)
	}

	// Delete pull request
	delete(c.pullRequests[repoFullName], number)

	// Delete associated labels
	if prLabels, exists := c.prLabels[repoFullName]; exists {
		delete(prLabels, number)
	}

	return nil
}

// ListPullRequests lists pull requests for a repository
func (c *Cache) ListPullRequests(ctx context.Context, repoFullName string, page, perPage int) ([]*models.PullRequest, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, 0, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Get pull requests for repository
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return []*models.PullRequest{}, 0, nil
	}

	// Convert map to slice
	total := len(repoMap)
	prs := make([]*models.PullRequest, 0, total)
	for _, pr := range repoMap {
		prs = append(prs, pr)
	}

	// Apply pagination
	start := (page - 1) * perPage
	if start >= total {
		return []*models.PullRequest{}, total, nil
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return prs[start:end], total, nil
}

// Issue operations

// AddIssue adds an issue to the cache
func (c *Cache) AddIssue(ctx context.Context, issue *models.Issue) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[issue.RepositoryFullName]; !exists {
		return fmt.Errorf("repository %s not found", issue.RepositoryFullName)
	}

	// Initialize map for repository if it doesn't exist
	if _, exists := c.issues[issue.RepositoryFullName]; !exists {
		c.issues[issue.RepositoryFullName] = make(map[int]*models.Issue)
	}

	// Check if issue already exists
	if _, exists := c.issues[issue.RepositoryFullName][issue.Number]; exists {
		return fmt.Errorf("issue %s#%d already exists", issue.RepositoryFullName, issue.Number)
	}

	// Add issue
	c.issues[issue.RepositoryFullName][issue.Number] = issue

	return nil
}

// GetIssue gets an issue from the cache
func (c *Cache) GetIssue(ctx context.Context, repoFullName string, number int) (*models.Issue, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return nil, fmt.Errorf("no issues found for repository %s", repoFullName)
	}

	issue, exists := repoMap[number]
	if !exists {
		return nil, fmt.Errorf("issue %s#%d not found", repoFullName, number)
	}

	return issue, nil
}

// UpdateIssue updates an issue in the cache
func (c *Cache) UpdateIssue(ctx context.Context, issue *models.Issue) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[issue.RepositoryFullName]; !exists {
		return fmt.Errorf("repository %s not found", issue.RepositoryFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[issue.RepositoryFullName]
	if !exists {
		return fmt.Errorf("no issues found for repository %s", issue.RepositoryFullName)
	}

	if _, exists := repoMap[issue.Number]; !exists {
		return fmt.Errorf("issue %s#%d not found", issue.RepositoryFullName, issue.Number)
	}

	// Update issue
	c.issues[issue.RepositoryFullName][issue.Number] = issue

	return nil
}

// DeleteIssue deletes an issue from the cache
func (c *Cache) DeleteIssue(ctx context.Context, repoFullName string, number int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return fmt.Errorf("no issues found for repository %s", repoFullName)
	}

	if _, exists := repoMap[number]; !exists {
		return fmt.Errorf("issue %s#%d not found", repoFullName, number)
	}

	// Delete issue
	delete(c.issues[repoFullName], number)

	// Delete associated labels
	if issueLabels, exists := c.issueLabels[repoFullName]; exists {
		delete(issueLabels, number)
	}

	return nil
}

// ListIssues lists issues for a repository
func (c *Cache) ListIssues(ctx context.Context, repoFullName string, page, perPage int) ([]*models.Issue, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, 0, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Get issues for repository
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return []*models.Issue{}, 0, nil
	}

	// Convert map to slice
	total := len(repoMap)
	issues := make([]*models.Issue, 0, total)
	for _, issue := range repoMap {
		issues = append(issues, issue)
	}

	// Apply pagination
	start := (page - 1) * perPage
	if start >= total {
		return []*models.Issue{}, total, nil
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return issues[start:end], total, nil
}

// Label operations

// AddLabel adds a label to the cache
func (c *Cache) AddLabel(ctx context.Context, label *models.Label) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if label already exists
	if _, exists := c.labels[label.Name]; exists {
		return fmt.Errorf("label %s already exists", label.Name)
	}

	// Add label
	c.labels[label.Name] = label

	return nil
}

// GetLabel gets a label from the cache
func (c *Cache) GetLabel(ctx context.Context, name string) (*models.Label, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if label exists
	label, exists := c.labels[name]
	if !exists {
		return nil, fmt.Errorf("label %s not found", name)
	}

	return label, nil
}

// UpdateLabel updates a label in the cache
func (c *Cache) UpdateLabel(ctx context.Context, label *models.Label) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if label exists
	if _, exists := c.labels[label.Name]; !exists {
		return fmt.Errorf("label %s not found", label.Name)
	}

	// Update label
	c.labels[label.Name] = label

	return nil
}

// DeleteLabel deletes a label from the cache
func (c *Cache) DeleteLabel(ctx context.Context, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if label exists
	if _, exists := c.labels[name]; !exists {
		return fmt.Errorf("label %s not found", name)
	}

	// Delete label
	delete(c.labels, name)

	// Delete label from all pull requests
	for repoName, prLabels := range c.prLabels {
		for prNumber, labels := range prLabels {
			delete(labels, name)
			if len(labels) == 0 {
				delete(prLabels, prNumber)
			}
		}
		if len(prLabels) == 0 {
			delete(c.prLabels, repoName)
		}
	}

	// Delete label from all issues
	for repoName, issueLabels := range c.issueLabels {
		for issueNumber, labels := range issueLabels {
			delete(labels, name)
			if len(labels) == 0 {
				delete(issueLabels, issueNumber)
			}
		}
		if len(issueLabels) == 0 {
			delete(c.issueLabels, repoName)
		}
	}

	return nil
}

// ListLabels lists all labels
func (c *Cache) ListLabels(ctx context.Context, page, perPage int) ([]*models.Label, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert map to slice
	total := len(c.labels)
	labels := make([]*models.Label, 0, total)
	for _, label := range c.labels {
		labels = append(labels, label)
	}

	// Apply pagination
	start := (page - 1) * perPage
	if start >= total {
		return []*models.Label{}, total, nil
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return labels[start:end], total, nil
}

// AddPullRequestLabel adds a label to a pull request
func (c *Cache) AddPullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return fmt.Errorf("no pull requests found for repository %s", repoFullName)
	}

	if _, exists := repoMap[prNumber]; !exists {
		return fmt.Errorf("pull request %s#%d not found", repoFullName, prNumber)
	}

	// Check if label exists
	if _, exists := c.labels[labelName]; !exists {
		return fmt.Errorf("label %s not found", labelName)
	}

	// Initialize maps if they don't exist
	if _, exists := c.prLabels[repoFullName]; !exists {
		c.prLabels[repoFullName] = make(map[int]map[string]struct{})
	}
	if _, exists := c.prLabels[repoFullName][prNumber]; !exists {
		c.prLabels[repoFullName][prNumber] = make(map[string]struct{})
	}

	// Add label to pull request
	c.prLabels[repoFullName][prNumber][labelName] = struct{}{}

	return nil
}

// RemovePullRequestLabel removes a label from a pull request
func (c *Cache) RemovePullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return fmt.Errorf("no pull requests found for repository %s", repoFullName)
	}

	if _, exists := repoMap[prNumber]; !exists {
		return fmt.Errorf("pull request %s#%d not found", repoFullName, prNumber)
	}

	// Check if label exists
	if _, exists := c.labels[labelName]; !exists {
		return fmt.Errorf("label %s not found", labelName)
	}

	// Check if pull request has labels
	if _, exists := c.prLabels[repoFullName]; !exists {
		return fmt.Errorf("no labels found for repository %s", repoFullName)
	}
	if _, exists := c.prLabels[repoFullName][prNumber]; !exists {
		return fmt.Errorf("no labels found for pull request %s#%d", repoFullName, prNumber)
	}

	// Remove label from pull request
	delete(c.prLabels[repoFullName][prNumber], labelName)

	// Clean up empty maps
	if len(c.prLabels[repoFullName][prNumber]) == 0 {
		delete(c.prLabels[repoFullName], prNumber)
	}
	if len(c.prLabels[repoFullName]) == 0 {
		delete(c.prLabels, repoFullName)
	}

	return nil
}

// ListPullRequestLabels lists labels for a pull request
func (c *Cache) ListPullRequestLabels(ctx context.Context, repoFullName string, prNumber int) ([]*models.Label, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if pull request exists
	repoMap, exists := c.pullRequests[repoFullName]
	if !exists {
		return nil, fmt.Errorf("no pull requests found for repository %s", repoFullName)
	}

	if _, exists := repoMap[prNumber]; !exists {
		return nil, fmt.Errorf("pull request %s#%d not found", repoFullName, prNumber)
	}

	// Check if pull request has labels
	if _, exists := c.prLabels[repoFullName]; !exists {
		return []*models.Label{}, nil
	}
	if _, exists := c.prLabels[repoFullName][prNumber]; !exists {
		return []*models.Label{}, nil
	}

	// Get labels for pull request
	labels := make([]*models.Label, 0, len(c.prLabels[repoFullName][prNumber]))
	for labelName := range c.prLabels[repoFullName][prNumber] {
		if label, exists := c.labels[labelName]; exists {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

// AddIssueLabel adds a label to an issue
func (c *Cache) AddIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return fmt.Errorf("no issues found for repository %s", repoFullName)
	}

	if _, exists := repoMap[issueNumber]; !exists {
		return fmt.Errorf("issue %s#%d not found", repoFullName, issueNumber)
	}

	// Check if label exists
	if _, exists := c.labels[labelName]; !exists {
		return fmt.Errorf("label %s not found", labelName)
	}

	// Initialize maps if they don't exist
	if _, exists := c.issueLabels[repoFullName]; !exists {
		c.issueLabels[repoFullName] = make(map[int]map[string]struct{})
	}
	if _, exists := c.issueLabels[repoFullName][issueNumber]; !exists {
		c.issueLabels[repoFullName][issueNumber] = make(map[string]struct{})
	}

	// Add label to issue
	c.issueLabels[repoFullName][issueNumber][labelName] = struct{}{}

	return nil
}

// RemoveIssueLabel removes a label from an issue
func (c *Cache) RemoveIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return fmt.Errorf("no issues found for repository %s", repoFullName)
	}

	if _, exists := repoMap[issueNumber]; !exists {
		return fmt.Errorf("issue %s#%d not found", repoFullName, issueNumber)
	}

	// Check if label exists
	if _, exists := c.labels[labelName]; !exists {
		return fmt.Errorf("label %s not found", labelName)
	}

	// Check if issue has labels
	if _, exists := c.issueLabels[repoFullName]; !exists {
		return fmt.Errorf("no labels found for repository %s", repoFullName)
	}
	if _, exists := c.issueLabels[repoFullName][issueNumber]; !exists {
		return fmt.Errorf("no labels found for issue %s#%d", repoFullName, issueNumber)
	}

	// Remove label from issue
	delete(c.issueLabels[repoFullName][issueNumber], labelName)

	// Clean up empty maps
	if len(c.issueLabels[repoFullName][issueNumber]) == 0 {
		delete(c.issueLabels[repoFullName], issueNumber)
	}
	if len(c.issueLabels[repoFullName]) == 0 {
		delete(c.issueLabels, repoFullName)
	}

	return nil
}

// ListIssueLabels lists labels for an issue
func (c *Cache) ListIssueLabels(ctx context.Context, repoFullName string, issueNumber int) ([]*models.Label, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if repository exists
	if _, exists := c.repositories[repoFullName]; !exists {
		return nil, fmt.Errorf("repository %s not found", repoFullName)
	}

	// Check if issue exists
	repoMap, exists := c.issues[repoFullName]
	if !exists {
		return nil, fmt.Errorf("no issues found for repository %s", repoFullName)
	}

	if _, exists := repoMap[issueNumber]; !exists {
		return nil, fmt.Errorf("issue %s#%d not found", repoFullName, issueNumber)
	}

	// Check if issue has labels
	if _, exists := c.issueLabels[repoFullName]; !exists {
		return []*models.Label{}, nil
	}
	if _, exists := c.issueLabels[repoFullName][issueNumber]; !exists {
		return []*models.Label{}, nil
	}

	// Get labels for issue
	labels := make([]*models.Label, 0, len(c.issueLabels[repoFullName][issueNumber]))
	for labelName := range c.issueLabels[repoFullName][issueNumber] {
		if label, exists := c.labels[labelName]; exists {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

// Close closes the cache
func (c *Cache) Close() error {
	return nil
}

// Ping checks if the cache is available
func (c *Cache) Ping(ctx context.Context) error {
	return nil
}

// Migrate performs any necessary migrations
func (c *Cache) Migrate(ctx context.Context) error {
	return nil
}
