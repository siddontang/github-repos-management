package github

import (
	"testing"
)

// TestCheckAuth tests the CheckAuth function
func TestCheckAuth(t *testing.T) {
	// Skip this test if running in CI environment
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Test authentication check
	err := CheckAuth()
	if err != nil {
		t.Logf("Authentication check failed: %v", err)
		t.Log("This test requires GitHub CLI to be authenticated. Run 'gh auth login' to authenticate.")
		t.Skip("GitHub CLI not authenticated, skipping test")
	}
}

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

// TestGetRepository tests the GetRepository function
func TestGetRepository(t *testing.T) {
	// Skip this test if running in CI environment
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check authentication first
	if err := CheckAuth(); err != nil {
		t.Skip("GitHub CLI not authenticated, skipping test")
	}

	// Create a client
	client := NewClient()

	// Test with a known public repository
	repo, err := client.GetRepository("pingcap", "tidb")
	if err != nil {
		t.Fatalf("GetRepository() error = %v", err)
	}
	if repo == nil {
		t.Fatal("GetRepository() returned nil repository")
	}
	if repo.Name != "tidb" {
		t.Errorf("GetRepository() repository name = %v, want %v", repo.Name, "tidb")
	}
	if repo.Owner.Login != "pingcap" {
		t.Errorf("GetRepository() repository owner = %v, want %v", repo.Owner.Login, "pingcap")
	}

	// Test with a non-existent repository
	_, err = client.GetRepository("this-user-does-not-exist", "this-repo-does-not-exist")
	if err == nil {
		t.Error("GetRepository() with non-existent repository should return an error")
	}
}

// TestListPullRequests tests the ListPullRequests function
func TestListPullRequests(t *testing.T) {
	// Skip this test if running in CI environment
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check authentication first
	if err := CheckAuth(); err != nil {
		t.Skip("GitHub CLI not authenticated, skipping test")
	}

	// Create a client
	client := NewClient()

	// Test with a known public repository
	options := &PullRequestOptions{
		State:   "all",
		PerPage: 5,
	}

	prs, err := client.ListPullRequests("pingcap", "tidb", options)
	if err != nil {
		t.Fatalf("ListPullRequests() error = %v", err)
	}

	// Verify we got some results
	if len(prs) == 0 {
		t.Log("No pull requests found, which is unusual for pingcap/tidb")
		t.Log("This might be due to API rate limiting or network issues")
	}

	// Verify we didn't get more than requested
	if len(prs) > options.PerPage {
		t.Errorf("ListPullRequests() returned %v pull requests, want at most %v", len(prs), options.PerPage)
	}

	// Test with a non-existent repository
	_, err = client.ListPullRequests("this-user-does-not-exist", "this-repo-does-not-exist", options)
	if err == nil {
		t.Error("ListPullRequests() with non-existent repository should return an error")
	}
}

// TestListIssues tests the ListIssues function
func TestListIssues(t *testing.T) {
	// Skip this test if running in CI environment
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check authentication first
	if err := CheckAuth(); err != nil {
		t.Skip("GitHub CLI not authenticated, skipping test")
	}

	// Create a client
	client := NewClient()

	// Test with a known public repository
	options := &IssueOptions{
		State:   "all",
		PerPage: 5,
	}

	issues, err := client.ListIssues("pingcap", "tidb", options)
	if err != nil {
		t.Fatalf("ListIssues() error = %v", err)
	}

	// Verify we got some results
	if len(issues) == 0 {
		t.Log("No issues found, which is unusual for pingcap/tidb")
		t.Log("This might be due to API rate limiting or network issues")
	}

	// Verify we didn't get more than requested
	if len(issues) > options.PerPage {
		t.Errorf("ListIssues() returned %v issues, want at most %v", len(issues), options.PerPage)
	}

	// Test with a non-existent repository
	_, err = client.ListIssues("this-user-does-not-exist", "this-repo-does-not-exist", options)
	if err == nil {
		t.Error("ListIssues() with non-existent repository should return an error")
	}
}

// TestGetRateLimit tests the GetRateLimit function
func TestGetRateLimit(t *testing.T) {
	// Skip this test if running in CI environment
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Check authentication first
	if err := CheckAuth(); err != nil {
		t.Skip("GitHub CLI not authenticated, skipping test")
	}

	// Create a client
	client := NewClient()

	// Test getting rate limit
	rateLimit, err := client.GetRateLimit()
	if err != nil {
		t.Fatalf("GetRateLimit() error = %v", err)
	}
	if rateLimit == nil {
		t.Fatal("GetRateLimit() returned nil rate limit")
	}
	if rateLimit.Limit <= 0 {
		t.Errorf("GetRateLimit() limit = %v, want > 0", rateLimit.Limit)
	}
	if rateLimit.Remaining < 0 {
		t.Errorf("GetRateLimit() remaining = %v, want >= 0", rateLimit.Remaining)
	}
}

// TestTruncate tests the truncate function
func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "Short string",
			input:  "Hello",
			maxLen: 10,
			want:   "Hello",
		},
		{
			name:   "Exact length",
			input:  "Hello",
			maxLen: 5,
			want:   "Hello",
		},
		{
			name:   "Long string",
			input:  "Hello, World!",
			maxLen: 8,
			want:   "Hello...",
		},
		{
			name:   "Empty string",
			input:  "",
			maxLen: 5,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}
