package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/siddontang/github-repos-management/internal/models"
)

// DB implements the db.DB interface with file-based persistence
type DB struct {
	sync.RWMutex

	// File path for persistence
	path string

	// In-memory data structures
	repositories map[string]*models.Repository
	pullRequests map[string]map[int]*models.PullRequest
	issues       map[string]map[int]*models.Issue
	labels       map[string]map[string]*models.Label

	// Relationships
	repoPRs     map[string][]int
	repoIssues  map[string][]int
	repoLabels  map[string]map[string]*models.Label
	prLabels    map[string]map[int][]string
	issueLabels map[string]map[int][]string
}

// data represents the structure for file persistence
type data struct {
	Repositories map[string]*models.Repository          `json:"repositories"`
	PullRequests map[string]map[int]*models.PullRequest `json:"pull_requests"`
	Issues       map[string]map[int]*models.Issue       `json:"issues"`
	Labels       map[string]map[string]*models.Label    `json:"labels"`
	RepoPRs      map[string][]int                       `json:"repo_prs"`
	RepoIssues   map[string][]int                       `json:"repo_issues"`
	RepoLabels   map[string]map[string]*models.Label    `json:"repo_labels"`
	PRLabels     map[string]map[int][]string            `json:"pr_labels"`
	IssueLabels  map[string]map[int][]string            `json:"issue_labels"`
}

// NewDB creates a new file-based database
func NewDB(path string) (*DB, error) {
	db := &DB{
		path:         path,
		repositories: make(map[string]*models.Repository),
		pullRequests: make(map[string]map[int]*models.PullRequest),
		issues:       make(map[string]map[int]*models.Issue),
		labels:       make(map[string]map[string]*models.Label),
		repoPRs:      make(map[string][]int),
		repoIssues:   make(map[string][]int),
		repoLabels:   make(map[string]map[string]*models.Label),
		prLabels:     make(map[string]map[int][]string),
		issueLabels:  make(map[string]map[int][]string),
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Load existing data if file exists
	if _, err := os.Stat(path); err == nil {
		if err := db.load(); err != nil {
			return nil, fmt.Errorf("failed to load data: %v", err)
		}
	}

	return db, nil
}

// load reads data from file
func (db *DB) load() error {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return err
	}

	var d data
	if err := json.Unmarshal(file, &d); err != nil {
		return err
	}

	db.repositories = d.Repositories
	db.pullRequests = d.PullRequests
	db.issues = d.Issues
	db.labels = d.Labels
	db.repoPRs = d.RepoPRs
	db.repoIssues = d.RepoIssues
	db.repoLabels = d.RepoLabels
	db.prLabels = d.PRLabels
	db.issueLabels = d.IssueLabels

	return nil
}

// sync writes data to file
func (db *DB) sync() error {
	d := data{
		Repositories: db.repositories,
		PullRequests: db.pullRequests,
		Issues:       db.issues,
		Labels:       db.labels,
		RepoPRs:      db.repoPRs,
		RepoIssues:   db.repoIssues,
		RepoLabels:   db.repoLabels,
		PRLabels:     db.prLabels,
		IssueLabels:  db.issueLabels,
	}

	file, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, file, 0644)
}

// Repository operations

// AddRepository adds a repository to the database
func (db *DB) AddRepository(ctx context.Context, repo *models.Repository) error {
	db.Lock()
	defer db.Unlock()

	db.repositories[repo.FullName] = repo
	return db.sync()
}

// GetRepository gets a repository from the database
func (db *DB) GetRepository(ctx context.Context, owner, name string) (*models.Repository, error) {
	db.RLock()
	defer db.RUnlock()

	fullName := owner + "/" + name
	repo, ok := db.repositories[fullName]
	if !ok {
		return nil, db.ErrRepositoryNotFound(fullName)
	}
	return repo, nil
}

// UpdateRepository updates a repository in the database
func (db *DB) UpdateRepository(ctx context.Context, repo *models.Repository) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.repositories[repo.FullName]; !ok {
		return db.ErrRepositoryNotFound(repo.FullName)
	}

	db.repositories[repo.FullName] = repo
	return db.sync()
}

// DeleteRepository deletes a repository from the database
func (db *DB) DeleteRepository(ctx context.Context, owner, name string) error {
	db.Lock()
	defer db.Unlock()

	fullName := owner + "/" + name
	if _, ok := db.repositories[fullName]; !ok {
		return db.ErrRepositoryNotFound(fullName)
	}

	delete(db.repositories, fullName)
	delete(db.pullRequests, fullName)
	delete(db.issues, fullName)
	delete(db.labels, fullName)
	delete(db.repoPRs, fullName)
	delete(db.repoIssues, fullName)
	delete(db.repoLabels, fullName)
	delete(db.prLabels, fullName)
	delete(db.issueLabels, fullName)

	return db.sync()
}

// ListRepositories lists repositories from the database
func (db *DB) ListRepositories(ctx context.Context, page, perPage int) ([]*models.Repository, int, error) {
	db.RLock()
	defer db.RUnlock()

	repos := make([]*models.Repository, 0, len(db.repositories))
	for _, repo := range db.repositories {
		repos = append(repos, repo)
	}

	total := len(repos)
	offset := (page - 1) * perPage
	if offset >= total {
		return []*models.Repository{}, total, nil
	}

	end := offset + perPage
	if end > total {
		end = total
	}

	return repos[offset:end], total, nil
}

// Pull Request operations

// AddPullRequest adds a pull request to the database
func (db *DB) AddPullRequest(ctx context.Context, pr *models.PullRequest) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.pullRequests[pr.RepositoryFullName]; !ok {
		db.pullRequests[pr.RepositoryFullName] = make(map[int]*models.PullRequest)
	}

	db.pullRequests[pr.RepositoryFullName][pr.Number] = pr

	if _, ok := db.repoPRs[pr.RepositoryFullName]; !ok {
		db.repoPRs[pr.RepositoryFullName] = make([]int, 0)
	}
	db.repoPRs[pr.RepositoryFullName] = append(db.repoPRs[pr.RepositoryFullName], pr.Number)

	return db.sync()
}

// GetPullRequest gets a pull request from the database
func (db *DB) GetPullRequest(ctx context.Context, repoFullName string, number int) (*models.PullRequest, error) {
	db.RLock()
	defer db.RUnlock()

	repoPRs, ok := db.pullRequests[repoFullName]
	if !ok {
		return nil, db.ErrPullRequestNotFound(repoFullName, number)
	}

	pr, ok := repoPRs[number]
	if !ok {
		return nil, db.ErrPullRequestNotFound(repoFullName, number)
	}

	return pr, nil
}

// ListPullRequests lists pull requests from the database
func (db *DB) ListPullRequests(ctx context.Context, repoFullName string, page, perPage int) ([]*models.PullRequest, int, error) {
	db.RLock()
	defer db.RUnlock()

	numbers, ok := db.repoPRs[repoFullName]
	if !ok {
		return []*models.PullRequest{}, 0, nil
	}

	total := len(numbers)
	offset := (page - 1) * perPage
	if offset >= total {
		return []*models.PullRequest{}, total, nil
	}

	end := offset + perPage
	if end > total {
		end = total
	}

	prs := make([]*models.PullRequest, 0, end-offset)
	for _, number := range numbers[offset:end] {
		if pr, ok := db.pullRequests[repoFullName][number]; ok {
			prs = append(prs, pr)
		}
	}

	return prs, total, nil
}

// UpdatePullRequest updates a pull request in the database
func (db *DB) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	// Just reuse the add method since it will overwrite
	return db.AddPullRequest(ctx, pr)
}

// DeletePullRequest deletes a pull request from the database
func (db *DB) DeletePullRequest(ctx context.Context, repoFullName string, number int) error {
	db.Lock()
	defer db.Unlock()

	repoPRs, ok := db.pullRequests[repoFullName]
	if !ok {
		return db.ErrPullRequestNotFound(repoFullName, number)
	}

	if _, ok := repoPRs[number]; !ok {
		return db.ErrPullRequestNotFound(repoFullName, number)
	}

	delete(repoPRs, number)

	// Remove from the list of PRs
	for i, n := range db.repoPRs[repoFullName] {
		if n == number {
			db.repoPRs[repoFullName] = append(db.repoPRs[repoFullName][:i], db.repoPRs[repoFullName][i+1:]...)
			break
		}
	}

	return db.sync()
}

// Issue operations

// AddIssue adds an issue to the database
func (db *DB) AddIssue(ctx context.Context, issue *models.Issue) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.issues[issue.RepositoryFullName]; !ok {
		db.issues[issue.RepositoryFullName] = make(map[int]*models.Issue)
	}

	db.issues[issue.RepositoryFullName][issue.Number] = issue

	if _, ok := db.repoIssues[issue.RepositoryFullName]; !ok {
		db.repoIssues[issue.RepositoryFullName] = make([]int, 0)
	}
	db.repoIssues[issue.RepositoryFullName] = append(db.repoIssues[issue.RepositoryFullName], issue.Number)

	return db.sync()
}

// GetIssue gets an issue from the database
func (db *DB) GetIssue(ctx context.Context, repoFullName string, number int) (*models.Issue, error) {
	db.RLock()
	defer db.RUnlock()

	repoIssues, ok := db.issues[repoFullName]
	if !ok {
		return nil, db.ErrIssueNotFound(repoFullName, number)
	}

	issue, ok := repoIssues[number]
	if !ok {
		return nil, db.ErrIssueNotFound(repoFullName, number)
	}

	return issue, nil
}

// ListIssues lists issues from the database
func (db *DB) ListIssues(ctx context.Context, repoFullName string, page, perPage int) ([]*models.Issue, int, error) {
	db.RLock()
	defer db.RUnlock()

	numbers, ok := db.repoIssues[repoFullName]
	if !ok {
		return []*models.Issue{}, 0, nil
	}

	total := len(numbers)
	offset := (page - 1) * perPage
	if offset >= total {
		return []*models.Issue{}, total, nil
	}

	end := offset + perPage
	if end > total {
		end = total
	}

	issues := make([]*models.Issue, 0, end-offset)
	for _, number := range numbers[offset:end] {
		if issue, ok := db.issues[repoFullName][number]; ok {
			issues = append(issues, issue)
		}
	}

	return issues, total, nil
}

// UpdateIssue updates an issue in the database
func (db *DB) UpdateIssue(ctx context.Context, issue *models.Issue) error {
	// Just reuse the add method since it will overwrite
	return db.AddIssue(ctx, issue)
}

// DeleteIssue deletes an issue from the database
func (db *DB) DeleteIssue(ctx context.Context, repoFullName string, number int) error {
	db.Lock()
	defer db.Unlock()

	repoIssues, ok := db.issues[repoFullName]
	if !ok {
		return db.ErrIssueNotFound(repoFullName, number)
	}

	if _, ok := repoIssues[number]; !ok {
		return db.ErrIssueNotFound(repoFullName, number)
	}

	delete(repoIssues, number)

	// Remove from the list of issues
	for i, n := range db.repoIssues[repoFullName] {
		if n == number {
			db.repoIssues[repoFullName] = append(db.repoIssues[repoFullName][:i], db.repoIssues[repoFullName][i+1:]...)
			break
		}
	}

	return db.sync()
}

// Label operations

// AddLabel adds a label to the database
func (db *DB) AddLabel(ctx context.Context, label *models.Label) error {
	db.Lock()
	defer db.Unlock()

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use the label's name as the repository name for now
	repoName := "global"

	if _, ok := db.labels[repoName]; !ok {
		db.labels[repoName] = make(map[string]*models.Label)
	}

	db.labels[repoName][label.Name] = label

	if _, ok := db.repoLabels[repoName]; !ok {
		db.repoLabels[repoName] = make(map[string]*models.Label)
	}
	db.repoLabels[repoName][label.Name] = label

	return db.sync()
}

// GetLabel gets a label from the database
func (db *DB) GetLabel(ctx context.Context, name string) (*models.Label, error) {
	db.RLock()
	defer db.RUnlock()

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use a global repository name for now
	repoName := "global"

	repoLabels, ok := db.labels[repoName]
	if !ok {
		return nil, db.ErrLabelNotFound(repoName, name)
	}

	label, ok := repoLabels[name]
	if !ok {
		return nil, db.ErrLabelNotFound(repoName, name)
	}

	return label, nil
}

// ListLabels lists labels from the database
func (db *DB) ListLabels(ctx context.Context, page, perPage int) ([]*models.Label, int, error) {
	db.RLock()
	defer db.RUnlock()

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use a global repository name for now
	repoName := "global"

	repoLabels, ok := db.repoLabels[repoName]
	if !ok {
		return []*models.Label{}, 0, nil
	}

	labels := make([]*models.Label, 0, len(repoLabels))
	for _, label := range repoLabels {
		labels = append(labels, label)
	}

	total := len(labels)
	offset := (page - 1) * perPage
	if offset >= total {
		return []*models.Label{}, total, nil
	}

	end := offset + perPage
	if end > total {
		end = total
	}

	return labels[offset:end], total, nil
}

// UpdateLabel updates a label in the database
func (db *DB) UpdateLabel(ctx context.Context, label *models.Label) error {
	// Just reuse the add method since it will overwrite
	return db.AddLabel(ctx, label)
}

// DeleteLabel deletes a label from the database
func (db *DB) DeleteLabel(ctx context.Context, name string) error {
	db.Lock()
	defer db.Unlock()

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use a global repository name for now
	repoName := "global"

	repoLabels, ok := db.labels[repoName]
	if !ok {
		return db.ErrLabelNotFound(repoName, name)
	}

	if _, ok := repoLabels[name]; !ok {
		return db.ErrLabelNotFound(repoName, name)
	}

	delete(repoLabels, name)
	delete(db.repoLabels[repoName], name)

	return db.sync()
}

// Pull request label operations

// AddPullRequestLabel adds a label to a pull request
func (db *DB) AddPullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.prLabels[repoFullName]; !ok {
		db.prLabels[repoFullName] = make(map[int][]string)
	}

	if _, ok := db.prLabels[repoFullName][prNumber]; !ok {
		db.prLabels[repoFullName][prNumber] = make([]string, 0)
	}

	// Check if the label already exists
	for _, name := range db.prLabels[repoFullName][prNumber] {
		if name == labelName {
			return nil
		}
	}

	db.prLabels[repoFullName][prNumber] = append(db.prLabels[repoFullName][prNumber], labelName)
	return db.sync()
}

// ListPullRequestLabels lists labels for a pull request
func (db *DB) ListPullRequestLabels(ctx context.Context, repoFullName string, prNumber int) ([]*models.Label, error) {
	db.RLock()
	defer db.RUnlock()

	if _, ok := db.prLabels[repoFullName]; !ok {
		return []*models.Label{}, nil
	}

	labelNames, ok := db.prLabels[repoFullName][prNumber]
	if !ok {
		return []*models.Label{}, nil
	}

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use a global repository name for now
	repoName := "global"

	labels := make([]*models.Label, 0, len(labelNames))
	for _, name := range labelNames {
		if label, ok := db.labels[repoName][name]; ok {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

// RemovePullRequestLabel removes a label from a pull request
func (db *DB) RemovePullRequestLabel(ctx context.Context, repoFullName string, prNumber int, labelName string) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.prLabels[repoFullName]; !ok {
		return nil
	}

	if _, ok := db.prLabels[repoFullName][prNumber]; !ok {
		return nil
	}

	// Find and remove the label
	for i, name := range db.prLabels[repoFullName][prNumber] {
		if name == labelName {
			db.prLabels[repoFullName][prNumber] = append(db.prLabels[repoFullName][prNumber][:i], db.prLabels[repoFullName][prNumber][i+1:]...)
			break
		}
	}

	return db.sync()
}

// Issue label operations

// AddIssueLabel adds a label to an issue
func (db *DB) AddIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.issueLabels[repoFullName]; !ok {
		db.issueLabels[repoFullName] = make(map[int][]string)
	}

	if _, ok := db.issueLabels[repoFullName][issueNumber]; !ok {
		db.issueLabels[repoFullName][issueNumber] = make([]string, 0)
	}

	// Check if the label already exists
	for _, name := range db.issueLabels[repoFullName][issueNumber] {
		if name == labelName {
			return nil
		}
	}

	db.issueLabels[repoFullName][issueNumber] = append(db.issueLabels[repoFullName][issueNumber], labelName)
	return db.sync()
}

// ListIssueLabels lists labels for an issue
func (db *DB) ListIssueLabels(ctx context.Context, repoFullName string, issueNumber int) ([]*models.Label, error) {
	db.RLock()
	defer db.RUnlock()

	if _, ok := db.issueLabels[repoFullName]; !ok {
		return []*models.Label{}, nil
	}

	labelNames, ok := db.issueLabels[repoFullName][issueNumber]
	if !ok {
		return []*models.Label{}, nil
	}

	// Since the Label struct doesn't have a RepositoryFullName field,
	// we'll use a global repository name for now
	repoName := "global"

	labels := make([]*models.Label, 0, len(labelNames))
	for _, name := range labelNames {
		if label, ok := db.labels[repoName][name]; ok {
			labels = append(labels, label)
		}
	}

	return labels, nil
}

// RemoveIssueLabel removes a label from an issue
func (db *DB) RemoveIssueLabel(ctx context.Context, repoFullName string, issueNumber int, labelName string) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.issueLabels[repoFullName]; !ok {
		return nil
	}

	if _, ok := db.issueLabels[repoFullName][issueNumber]; !ok {
		return nil
	}

	// Find and remove the label
	for i, name := range db.issueLabels[repoFullName][issueNumber] {
		if name == labelName {
			db.issueLabels[repoFullName][issueNumber] = append(db.issueLabels[repoFullName][issueNumber][:i], db.issueLabels[repoFullName][issueNumber][i+1:]...)
			break
		}
	}

	return db.sync()
}

// Maintenance operations

// Close closes the database
func (db *DB) Close() error {
	// Sync any pending changes
	return db.sync()
}

// Ping checks if the database is available
func (db *DB) Ping(ctx context.Context) error {
	// The file DB is always available if we can access the file
	_, err := os.Stat(db.path)
	return err
}

// Sync syncs the database to disk
func (db *DB) Sync() error {
	db.Lock()
	defer db.Unlock()

	return db.sync()
}

// Error helpers

func (db *DB) ErrRepositoryNotFound(fullName string) error {
	return fmt.Errorf("repository %s not found", fullName)
}

func (db *DB) ErrPullRequestNotFound(fullName string, number int) error {
	return fmt.Errorf("pull request %d not found in repository %s", number, fullName)
}

func (db *DB) ErrIssueNotFound(fullName string, number int) error {
	return fmt.Errorf("issue %d not found in repository %s", number, fullName)
}

func (db *DB) ErrLabelNotFound(fullName string, name string) error {
	return fmt.Errorf("label %s not found in repository %s", name, fullName)
}
