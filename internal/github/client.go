package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Client represents a GitHub client that uses the gh CLI
type Client struct {
	// Add any client-specific configuration here
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)

// NewClient creates a new GitHub client
func NewClient() *Client {
	return &Client{}
}

// CheckAuth checks if the user is authenticated with GitHub
func CheckAuth() error {
	cmd := exec.Command("gh", "auth", "status")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub authentication failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// Login performs GitHub authentication
func Login() error {
	cmd := exec.Command("gh", "auth", "login")
	cmd.Stdin = strings.NewReader("\n") // Default options
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub login failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// GetRepository gets information about a repository
func (c *Client) GetRepository(owner, name string) (*Repository, error) {
	// Build the command to use gh repo view
	args := []string{"repo", "view", fmt.Sprintf("%s/%s", owner, name), "--json", "name,owner,nameWithOwner,description,url,homepageUrl,isPrivate,createdAt,updatedAt"}
	cmdStr := fmt.Sprintf("gh %s", strings.Join(args, " "))
	fmt.Printf("Executing command: %s\n", cmdStr)

	// Execute the command
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
		return nil, fmt.Errorf("failed to get repository: %w, stderr: %s", err, stderr.String())
	}

	// Print the output for debugging
	fmt.Printf("Command output: %s\n", stdout.String())

	// Parse the JSON output
	var ghRepo struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		NameWithOwner string `json:"nameWithOwner"`
		Description   string `json:"description"`
		URL           string `json:"url"`
		HomepageURL   string `json:"homepageUrl"`
		IsPrivate     bool   `json:"isPrivate"`
		CreatedAt     string `json:"createdAt"`
		UpdatedAt     string `json:"updatedAt"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &ghRepo); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		fmt.Printf("JSON content: %s\n", stdout.String())
		return nil, fmt.Errorf("failed to parse repository data: %w", err)
	}

	// Parse dates
	createdAt, err := time.Parse(time.RFC3339, ghRepo.CreatedAt)
	if err != nil {
		fmt.Printf("Failed to parse createdAt date: %v\n", err)
		createdAt = time.Now() // Use current time as fallback
	}

	updatedAt, err := time.Parse(time.RFC3339, ghRepo.UpdatedAt)
	if err != nil {
		fmt.Printf("Failed to parse updatedAt date: %v\n", err)
		updatedAt = time.Now() // Use current time as fallback
	}

	// Create repository
	repository := &Repository{
		Owner:       User{Login: ghRepo.Owner.Login},
		Name:        ghRepo.Name,
		FullName:    ghRepo.NameWithOwner,
		Description: ghRepo.Description,
		URL:         ghRepo.URL,
		HTMLURL:     ghRepo.HomepageURL,
		Private:     ghRepo.IsPrivate,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	fmt.Printf("Repository object created: %+v\n", repository)
	return repository, nil
}

// ListPullRequests lists pull requests for a repository
func (c *Client) ListPullRequests(owner, name string, options *PullRequestOptions) ([]*PullRequest, error) {
	// Build the command to use gh pr list
	args := []string{"pr", "list", "--repo", fmt.Sprintf("%s/%s", owner, name), "--json", "number,title,state,author,createdAt,updatedAt,url"}

	// Add query parameters
	if options != nil {
		if options.State != "" {
			args = append(args, "--state", options.State)
		}
		if options.PerPage > 0 {
			args = append(args, "--limit", strconv.Itoa(options.PerPage))
		}
	}

	cmdStr := fmt.Sprintf("gh %s", strings.Join(args, " "))
	fmt.Printf("Executing command: %s\n", cmdStr)

	// Execute the command
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
		return nil, fmt.Errorf("failed to list pull requests: %w, stderr: %s", err, stderr.String())
	}

	// Print the output for debugging
	fmt.Printf("Command output length: %d bytes\n", len(stdout.String()))
	if len(stdout.String()) < 1000 {
		fmt.Printf("Command output: %s\n", stdout.String())
	}

	// Parse the JSON output
	var ghPRs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		URL       string `json:"url"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &ghPRs); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		fmt.Printf("JSON content (first 200 chars): %s\n", truncate(stdout.String(), 200))
		return nil, fmt.Errorf("failed to parse pull requests data: %w", err)
	}

	// Convert to our model
	prs := make([]*PullRequest, 0, len(ghPRs))
	for _, ghPR := range ghPRs {
		// Parse dates
		createdAt, err := time.Parse(time.RFC3339, ghPR.CreatedAt)
		if err != nil {
			fmt.Printf("Failed to parse createdAt date: %v\n", err)
			createdAt = time.Now() // Use current time as fallback
		}

		updatedAt, err := time.Parse(time.RFC3339, ghPR.UpdatedAt)
		if err != nil {
			fmt.Printf("Failed to parse updatedAt date: %v\n", err)
			updatedAt = time.Now() // Use current time as fallback
		}

		pr := &PullRequest{
			Number:    ghPR.Number,
			Title:     ghPR.Title,
			State:     ghPR.State,
			User:      User{Login: ghPR.Author.Login},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			HTMLURL:   ghPR.URL,
		}
		prs = append(prs, pr)
	}

	fmt.Printf("Parsed %d pull requests\n", len(prs))
	return prs, nil
}

// ListIssues lists issues for a repository
func (c *Client) ListIssues(owner, name string, options *IssueOptions) ([]*Issue, error) {
	// Build the command to use gh issue list
	args := []string{"issue", "list", "--repo", fmt.Sprintf("%s/%s", owner, name), "--json", "number,title,state,author,createdAt,updatedAt,url"}

	// Add query parameters
	if options != nil {
		if options.State != "" {
			args = append(args, "--state", options.State)
		}
		if options.PerPage > 0 {
			args = append(args, "--limit", strconv.Itoa(options.PerPage))
		}
	}

	cmdStr := fmt.Sprintf("gh %s", strings.Join(args, " "))
	fmt.Printf("Executing command: %s\n", cmdStr)

	// Execute the command
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
		return nil, fmt.Errorf("failed to list issues: %w, stderr: %s", err, stderr.String())
	}

	// Print the output for debugging
	fmt.Printf("Command output length: %d bytes\n", len(stdout.String()))
	if len(stdout.String()) < 1000 {
		fmt.Printf("Command output: %s\n", stdout.String())
	}

	// Parse the JSON output
	var ghIssues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		URL       string `json:"url"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &ghIssues); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		fmt.Printf("JSON content (first 200 chars): %s\n", truncate(stdout.String(), 200))
		return nil, fmt.Errorf("failed to parse issues data: %w", err)
	}

	// Convert to our model
	issues := make([]*Issue, 0, len(ghIssues))
	for _, ghIssue := range ghIssues {
		// Parse dates
		createdAt, err := time.Parse(time.RFC3339, ghIssue.CreatedAt)
		if err != nil {
			fmt.Printf("Failed to parse createdAt date: %v\n", err)
			createdAt = time.Now() // Use current time as fallback
		}

		updatedAt, err := time.Parse(time.RFC3339, ghIssue.UpdatedAt)
		if err != nil {
			fmt.Printf("Failed to parse updatedAt date: %v\n", err)
			updatedAt = time.Now() // Use current time as fallback
		}

		issue := &Issue{
			Number:    ghIssue.Number,
			Title:     ghIssue.Title,
			State:     ghIssue.State,
			User:      User{Login: ghIssue.Author.Login},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			HTMLURL:   ghIssue.URL,
		}
		issues = append(issues, issue)
	}

	fmt.Printf("Parsed %d issues\n", len(issues))
	return issues, nil
}

// Helper function to truncate a string
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetRateLimit gets the current GitHub API rate limit
func (c *Client) GetRateLimit() (*RateLimit, error) {
	// Build the command
	args := []string{"api", "rate_limit"}

	// Execute the command
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w, stderr: %s", err, stderr.String())
	}

	// Parse the JSON output
	var response struct {
		Resources struct {
			Core RateLimit `json:"core"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse rate limit data: %w", err)
	}

	// Set reset time
	response.Resources.Core.ResetTime = time.Unix(response.Resources.Core.Reset, 0)

	return &response.Resources.Core, nil
}
