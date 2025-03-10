package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/siddontang/github-repos-management/internal/cache"
	"github.com/siddontang/github-repos-management/internal/cache/memory"
	"github.com/siddontang/github-repos-management/internal/config"
	"github.com/siddontang/github-repos-management/internal/github"
	"github.com/siddontang/github-repos-management/internal/models"
)

// Service represents the main service for the GitHub repository management
type Service struct {
	config     *config.Config
	cache      cache.Cache
	ghClient   github.ClientInterface
	syncMutex  sync.Mutex
	syncStatus map[string]string // repository full name -> status
	startTime  time.Time
}

// NewService creates a new service instance
func NewService(cfg *config.Config) (*Service, error) {
	// Create GitHub client
	ghClient := github.NewClient()

	// Create cache provider based on configuration
	var cacheProvider cache.Provider
	switch cfg.Database.Type {
	case "sqlite":
		// TODO: Implement SQLite provider
		cacheProvider = memory.NewProvider() // Use memory cache for now
	case "mysql":
		// TODO: Implement MySQL provider
		cacheProvider = memory.NewProvider() // Use memory cache for now
	default:
		cacheProvider = memory.NewProvider()
	}

	// Create cache instance
	cacheInstance, err := cacheProvider(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &Service{
		config:     cfg,
		cache:      cacheInstance,
		ghClient:   ghClient,
		syncStatus: make(map[string]string),
		startTime:  time.Now(),
	}, nil
}

// Close closes the service and its resources
func (s *Service) Close() error {
	return s.cache.Close()
}

// Repository operations

// AddRepository adds a new repository to be tracked
func (s *Service) AddRepository(ctx context.Context, fullName string) (*models.Repository, error) {
	// Parse owner and name
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return nil, ErrInvalidRepositoryName
	}
	owner, name := parts[0], parts[1]

	// Check if repository already exists
	existingRepo, err := s.cache.GetRepository(ctx, owner, name)
	if err == nil && existingRepo != nil {
		log.Printf("Repository %s already exists in cache", fullName)
		return existingRepo, nil
	}

	log.Printf("Adding new repository: %s", fullName)

	// Get repository from GitHub
	ghRepo, err := s.ghClient.GetRepository(owner, name)
	if err != nil {
		log.Printf("Error fetching repository from GitHub: %v", err)
		return nil, fmt.Errorf("failed to get repository from GitHub: %w", err)
	}

	log.Printf("Successfully fetched repository from GitHub: %s", fullName)

	// Create repository model
	repo := &models.Repository{
		Owner:        ghRepo.Owner.Login,
		Name:         ghRepo.Name,
		FullName:     ghRepo.FullName,
		Description:  ghRepo.Description,
		URL:          ghRepo.URL,
		HTMLURL:      ghRepo.HTMLURL,
		IsPrivate:    ghRepo.Private,
		LastSyncedAt: time.Now(), // Set initial sync time
		CreatedAt:    ghRepo.CreatedAt,
		UpdatedAt:    ghRepo.UpdatedAt,
	}

	// Add repository to cache
	if err := s.cache.AddRepository(ctx, repo); err != nil {
		log.Printf("Error adding repository to cache: %v", err)
		return nil, fmt.Errorf("failed to add repository to cache: %w", err)
	}

	log.Printf("Successfully added repository to cache: %s", fullName)

	// Start background sync
	go func() {
		log.Printf("Starting background sync for repository: %s", fullName)
		if err := s.syncRepository(context.Background(), owner, name); err != nil {
			log.Printf("Error syncing repository %s: %v", fullName, err)
		} else {
			log.Printf("Successfully synced repository: %s", fullName)
		}
	}()

	return repo, nil
}

// GetRepository gets a repository by owner and name
func (s *Service) GetRepository(ctx context.Context, owner, name string) (*models.Repository, error) {
	repo, err := s.cache.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, ErrRepositoryNotFound
	}
	return repo, nil
}

// ListRepositories lists all tracked repositories
func (s *Service) ListRepositories(ctx context.Context, page, perPage int) ([]*models.Repository, int, error) {
	return s.cache.ListRepositories(ctx, page, perPage)
}

// DeleteRepository removes a repository from tracking
func (s *Service) DeleteRepository(ctx context.Context, owner, name string) error {
	err := s.cache.DeleteRepository(ctx, owner, name)
	if err != nil {
		return ErrRepositoryNotFound
	}
	return nil
}

// RefreshRepository forces a refresh of repository data
func (s *Service) RefreshRepository(ctx context.Context, owner, name string) error {
	// Check if repository exists
	_, err := s.cache.GetRepository(ctx, owner, name)
	if err != nil {
		return ErrRepositoryNotFound
	}

	// Start sync in background
	go func() {
		syncCtx := context.Background()
		if err := s.syncRepository(syncCtx, owner, name); err != nil {
			// Log the error but don't return it since we're in a goroutine
			fmt.Printf("Error refreshing repository %s/%s: %v\n", owner, name, err)
		}
	}()

	return nil
}

// syncRepository syncs a repository's data from GitHub
func (s *Service) syncRepository(ctx context.Context, owner, name string) error {
	fullName := fmt.Sprintf("%s/%s", owner, name)

	// Set sync status
	s.syncMutex.Lock()
	s.syncStatus[fullName] = "syncing"
	s.syncMutex.Unlock()

	// Ensure status is updated when done
	defer func() {
		s.syncMutex.Lock()
		delete(s.syncStatus, fullName)
		s.syncMutex.Unlock()
	}()

	// Get repository from cache
	repo, err := s.cache.GetRepository(ctx, owner, name)
	if err != nil {
		s.syncMutex.Lock()
		s.syncStatus[fullName] = fmt.Sprintf("error: %v", err)
		s.syncMutex.Unlock()
		return fmt.Errorf("repository not found: %w", err)
	}

	// Sync pull requests
	if err := s.syncPullRequests(ctx, owner, name); err != nil {
		s.syncMutex.Lock()
		s.syncStatus[fullName] = fmt.Sprintf("error syncing pull requests: %v", err)
		s.syncMutex.Unlock()
		return fmt.Errorf("failed to sync pull requests: %w", err)
	}

	// Sync issues
	if err := s.syncIssues(ctx, owner, name); err != nil {
		s.syncMutex.Lock()
		s.syncStatus[fullName] = fmt.Sprintf("error syncing issues: %v", err)
		s.syncMutex.Unlock()
		return fmt.Errorf("failed to sync issues: %w", err)
	}

	// Update last synced time after successful sync
	repo.LastSyncedAt = time.Now()
	if err := s.cache.UpdateRepository(ctx, repo); err != nil {
		return fmt.Errorf("failed to update last synced time: %w", err)
	}

	return nil
}

// syncPullRequests syncs pull requests for a repository
func (s *Service) syncPullRequests(ctx context.Context, owner, name string) error {
	// Get repository
	repo, err := s.cache.GetRepository(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Get pull requests from GitHub
	options := &github.PullRequestOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		PerPage:   100,
		Page:      1,
	}

	prs, err := s.ghClient.ListPullRequests(owner, name, options)
	if err != nil {
		return fmt.Errorf("failed to list pull requests: %w", err)
	}

	// Process pull requests
	for _, ghPR := range prs {
		// Create pull request model
		pr := &models.PullRequest{
			RepositoryFullName: repo.FullName,
			Number:             ghPR.Number,
			Title:              ghPR.Title,
			Body:               ghPR.Body,
			State:              ghPR.State,
			URL:                ghPR.URL,
			HTMLURL:            ghPR.HTMLURL,
			UserLogin:          ghPR.User.Login,
			UserAvatarURL:      ghPR.User.AvatarURL,
			UserURL:            ghPR.User.URL,
			UserHTMLURL:        ghPR.User.HTMLURL,
			CreatedAt:          ghPR.CreatedAt,
			UpdatedAt:          ghPR.UpdatedAt,
			ClosedAt:           ghPR.ClosedAt,
			MergedAt:           ghPR.MergedAt,
		}

		// Check if pull request exists
		existingPR, err := s.cache.GetPullRequest(ctx, repo.FullName, ghPR.Number)
		if err == nil && existingPR != nil {
			// Update existing pull request
			if err := s.cache.UpdatePullRequest(ctx, pr); err != nil {
				continue
			}
		} else {
			// Add new pull request
			if err := s.cache.AddPullRequest(ctx, pr); err != nil {
				continue
			}
		}

		// Process labels
		for _, ghLabel := range ghPR.Labels {
			// Create label model
			label := &models.Label{
				Name:        ghLabel.Name,
				Color:       ghLabel.Color,
				Description: ghLabel.Description,
			}

			// Check if label exists
			existingLabel, err := s.cache.GetLabel(ctx, ghLabel.Name)
			if err != nil || existingLabel == nil {
				// Add new label
				if err := s.cache.AddLabel(ctx, label); err != nil {
					continue
				}
			}

			// Add label to pull request
			if err := s.cache.AddPullRequestLabel(ctx, repo.FullName, ghPR.Number, ghLabel.Name); err != nil {
				// Ignore errors
			}
		}
	}

	return nil
}

// syncIssues syncs issues for a repository
func (s *Service) syncIssues(ctx context.Context, owner, name string) error {
	// Get repository
	repo, err := s.cache.GetRepository(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Get issues from GitHub
	options := &github.IssueOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		PerPage:   100,
		Page:      1,
	}

	issues, err := s.ghClient.ListIssues(owner, name, options)
	if err != nil {
		return fmt.Errorf("failed to list issues: %w", err)
	}

	// Process issues
	for _, ghIssue := range issues {
		// Create issue model
		issue := &models.Issue{
			RepositoryFullName: repo.FullName,
			Number:             ghIssue.Number,
			Title:              ghIssue.Title,
			Body:               ghIssue.Body,
			State:              ghIssue.State,
			URL:                ghIssue.URL,
			HTMLURL:            ghIssue.HTMLURL,
			UserLogin:          ghIssue.User.Login,
			UserAvatarURL:      ghIssue.User.AvatarURL,
			UserURL:            ghIssue.User.URL,
			UserHTMLURL:        ghIssue.User.HTMLURL,
			CreatedAt:          ghIssue.CreatedAt,
			UpdatedAt:          ghIssue.UpdatedAt,
			ClosedAt:           ghIssue.ClosedAt,
		}

		// Check if issue exists
		existingIssue, err := s.cache.GetIssue(ctx, repo.FullName, ghIssue.Number)
		if err == nil && existingIssue != nil {
			// Update existing issue
			if err := s.cache.UpdateIssue(ctx, issue); err != nil {
				continue
			}
		} else {
			// Add new issue
			if err := s.cache.AddIssue(ctx, issue); err != nil {
				continue
			}
		}

		// Process labels
		for _, ghLabel := range ghIssue.Labels {
			// Create label model
			label := &models.Label{
				Name:        ghLabel.Name,
				Color:       ghLabel.Color,
				Description: ghLabel.Description,
			}

			// Check if label exists
			existingLabel, err := s.cache.GetLabel(ctx, ghLabel.Name)
			if err != nil || existingLabel == nil {
				// Add new label
				if err := s.cache.AddLabel(ctx, label); err != nil {
					continue
				}
			}

			// Add label to issue
			if err := s.cache.AddIssueLabel(ctx, repo.FullName, ghIssue.Number, ghLabel.Name); err != nil {
				// Ignore errors
			}
		}
	}

	return nil
}

// Pull request operations

// ListPullRequests lists pull requests for a repository or across all repositories
func (s *Service) ListPullRequests(ctx context.Context, filter *models.PullRequestFilter) ([]*models.PullRequest, *models.Pagination, error) {
	return s.listAllPullRequests(ctx, filter)
}

// listAllPullRequests lists pull requests across all repositories or for a specific repository
func (s *Service) listAllPullRequests(ctx context.Context, filter *models.PullRequestFilter) ([]*models.PullRequest, *models.Pagination, error) {
	// Get repositories to process
	var repos []*models.Repository
	var err error

	// If a specific repository is requested
	if filter.Repo != "" {
		// Parse repository owner and name
		parts := strings.Split(filter.Repo, "/")
		if len(parts) != 2 {
			return nil, nil, ErrInvalidRepositoryName
		}
		owner, name := parts[0], parts[1]

		// Get the specific repository
		repo, err := s.cache.GetRepository(ctx, owner, name)
		if err != nil {
			return nil, nil, ErrRepositoryNotFound
		}
		repos = []*models.Repository{repo}
	} else {
		// Get all repositories
		repos, _, err = s.cache.ListRepositories(ctx, 1, 1000) // Assuming we won't have more than 1000 repos
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list repositories: %w", err)
		}
	}

	// Collect all pull requests
	var allPRs []*models.PullRequest
	for _, repo := range repos {
		prs, _, err := s.cache.ListPullRequests(ctx, repo.FullName, 1, 1000) // Get all PRs, we'll paginate later
		if err != nil {
			// Log error but continue
			continue
		}
		allPRs = append(allPRs, prs...)
	}

	// Apply filters
	var filteredPRs []*models.PullRequest
	for _, pr := range allPRs {
		// Filter by state (case-insensitive comparison)
		if filter.State != "" && !strings.EqualFold(pr.State, filter.State) {
			continue
		}

		// Filter by author
		if filter.Author != "" && !strings.EqualFold(pr.UserLogin, filter.Author) {
			continue
		}

		// Filter by label (would need to fetch labels for each PR)
		// This is simplified - in a real implementation, you'd need to check labels

		// Add to filtered list
		filteredPRs = append(filteredPRs, pr)
	}

	// Sort the PRs (simplified - in a real implementation, you'd need more complex sorting)
	// For now, just sort by creation date
	sort.Slice(filteredPRs, func(i, j int) bool {
		if filter.Direction == "asc" {
			return filteredPRs[i].CreatedAt.Before(filteredPRs[j].CreatedAt)
		}
		return filteredPRs[i].CreatedAt.After(filteredPRs[j].CreatedAt)
	})

	// Apply pagination
	total := len(filteredPRs)
	start := (filter.Page - 1) * filter.PerPage
	if start >= total {
		return []*models.PullRequest{}, &models.Pagination{
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			Total:      total,
			TotalPages: (total + filter.PerPage - 1) / filter.PerPage,
		}, nil
	}

	end := start + filter.PerPage
	if end > total {
		end = total
	}

	// Create pagination
	pagination := &models.Pagination{
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		Total:      total,
		TotalPages: (total + filter.PerPage - 1) / filter.PerPage,
	}

	return filteredPRs[start:end], pagination, nil
}

// Issue operations

// ListIssues lists issues for a repository or across all repositories
func (s *Service) ListIssues(ctx context.Context, filter *models.IssueFilter) ([]*models.Issue, *models.Pagination, error) {
	return s.listAllIssues(ctx, filter)
}

// listAllIssues lists issues across all repositories or for a specific repository
func (s *Service) listAllIssues(ctx context.Context, filter *models.IssueFilter) ([]*models.Issue, *models.Pagination, error) {
	// Get repositories to process
	var repos []*models.Repository
	var err error

	// If a specific repository is requested
	if filter.Repo != "" {
		// Parse repository owner and name
		parts := strings.Split(filter.Repo, "/")
		if len(parts) != 2 {
			return nil, nil, ErrInvalidRepositoryName
		}
		owner, name := parts[0], parts[1]

		// Get the specific repository
		repo, err := s.cache.GetRepository(ctx, owner, name)
		if err != nil {
			return nil, nil, ErrRepositoryNotFound
		}
		repos = []*models.Repository{repo}
	} else {
		// Get all repositories
		repos, _, err = s.cache.ListRepositories(ctx, 1, 1000) // Assuming we won't have more than 1000 repos
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list repositories: %w", err)
		}
	}

	// Collect all issues
	var allIssues []*models.Issue
	for _, repo := range repos {
		issues, _, err := s.cache.ListIssues(ctx, repo.FullName, 1, 1000) // Get all issues, we'll paginate later
		if err != nil {
			// Log error but continue
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	// Apply filters
	var filteredIssues []*models.Issue
	for _, issue := range allIssues {
		// Filter by state (case-insensitive comparison)
		if filter.State != "" && !strings.EqualFold(issue.State, filter.State) {
			continue
		}

		// Filter by author
		if filter.Author != "" && !strings.EqualFold(issue.UserLogin, filter.Author) {
			continue
		}

		// Filter by label (would need to fetch labels for each issue)
		// This is simplified - in a real implementation, you'd need to check labels

		// Add to filtered list
		filteredIssues = append(filteredIssues, issue)
	}

	// Sort the issues (simplified - in a real implementation, you'd need more complex sorting)
	// For now, just sort by creation date
	sort.Slice(filteredIssues, func(i, j int) bool {
		if filter.Direction == "asc" {
			return filteredIssues[i].CreatedAt.Before(filteredIssues[j].CreatedAt)
		}
		return filteredIssues[i].CreatedAt.After(filteredIssues[j].CreatedAt)
	})

	// Apply pagination
	total := len(filteredIssues)
	start := (filter.Page - 1) * filter.PerPage
	if start >= total {
		return []*models.Issue{}, &models.Pagination{
			Page:       filter.Page,
			PerPage:    filter.PerPage,
			Total:      total,
			TotalPages: (total + filter.PerPage - 1) / filter.PerPage,
		}, nil
	}

	end := start + filter.PerPage
	if end > total {
		end = total
	}

	// Create pagination
	pagination := &models.Pagination{
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		Total:      total,
		TotalPages: (total + filter.PerPage - 1) / filter.PerPage,
	}

	return filteredIssues[start:end], pagination, nil
}

// Service operations

// RefreshAll forces a refresh of all repository data
func (s *Service) RefreshAll(ctx context.Context) error {
	// Get all repositories
	repos, _, err := s.cache.ListRepositories(ctx, 1, 1000) // Assuming we won't have more than 1000 repos
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Refresh each repository
	for _, repo := range repos {
		go func(owner, name string) {
			syncCtx := context.Background()
			if err := s.syncRepository(syncCtx, owner, name); err != nil {
				// Log the error but don't return it since we're in a goroutine
				fmt.Printf("Error refreshing repository %s/%s: %v\n", owner, name, err)
			}
		}(repo.Owner, repo.Name)
	}

	return nil
}

// GetStatus returns the current status of the service
func (s *Service) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	// Get all repositories
	repos, total, err := s.cache.ListRepositories(ctx, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Count syncing and error repositories
	s.syncMutex.Lock()
	syncing := len(s.syncStatus)
	errors := 0
	for _, status := range s.syncStatus {
		if strings.HasPrefix(status, "error") {
			errors++
		}
	}
	s.syncMutex.Unlock()

	// Get rate limit
	rateLimit, err := s.ghClient.GetRateLimit()
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}

	// Find last sync time
	var lastSync time.Time
	for _, repo := range repos {
		if repo.LastSyncedAt.After(lastSync) {
			lastSync = repo.LastSyncedAt
		}
	}

	// Build status
	status := map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"uptime":  int(time.Since(s.startTime).Seconds()),
		"repositories": map[string]interface{}{
			"total":   total,
			"syncing": syncing,
			"error":   errors,
		},
		"last_sync": lastSync,
		"github_rate_limit": map[string]interface{}{
			"limit":     rateLimit.Limit,
			"remaining": rateLimit.Remaining,
			"reset_at":  time.Unix(rateLimit.Reset, 0),
		},
	}

	return status, nil
}
