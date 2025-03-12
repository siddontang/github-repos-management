package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

func main() {
	// Root command
	rootCmd := &cobra.Command{
		Use:   "ghrepos",
		Short: "GitHub Repository Management CLI",
		Long:  "A CLI tool to manage and track GitHub repositories directly",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// No need to initialize client here as each command creates its own client
			cmd.SetContext(cmd.Context())
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Repository command
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage tracked repositories",
		Long:  "Track, untrack, and list GitHub repositories directly",
	}

	// Add repository command
	addRepoCmd := &cobra.Command{
		Use:   "add [owner/name]",
		Short: "Add a repository to track",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			repo, err := client.AddRepository(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error adding repository: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Repository %s added successfully\n", repo.FullName)
		},
	}

	// List repositories command
	listRepoCmd := &cobra.Command{
		Use:   "list",
		Short: "List tracked repositories",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			page, _ := cmd.Flags().GetInt("page")
			perPage, _ := cmd.Flags().GetInt("per-page")

			resp, err := client.ListRepositories(page, perPage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing repositories: %v\n", err)
				os.Exit(1)
			}

			// Print repositories
			fmt.Printf("%-40s %-20s %-20s %s\n", "REPOSITORY", "PRIVATE", "LAST SYNCED", "URL")
			for _, repo := range resp.Data {
				lastSynced := repo.LastSyncedAt.Format("2006-01-02 15:04:05")
				isPrivate := "No"
				if repo.IsPrivate {
					isPrivate = "Yes"
				}
				fmt.Printf("%-40s %-20s %-20s %s\n", repo.FullName, isPrivate, lastSynced, repo.HTMLURL)
			}

			// Print pagination info
			fmt.Printf("\nPage %d of %d (Total: %d)\n", resp.Pagination.Page, resp.Pagination.TotalPages, resp.Pagination.Total)
		},
	}
	listRepoCmd.Flags().IntP("page", "p", 1, "Page number")
	listRepoCmd.Flags().IntP("per-page", "n", 10, "Items per page")

	// Remove repository command
	removeRepoCmd := &cobra.Command{
		Use:   "remove [owner/name]",
		Short: "Remove a repository from tracking",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			parts := strings.Split(args[0], "/")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid repository name format, expected 'owner/repo'\n")
				os.Exit(1)
			}
			owner, name := parts[0], parts[1]

			err = client.RemoveRepository(owner, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error removing repository: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Repository %s removed successfully\n", args[0])
		},
	}

	// Refresh repository command
	refreshRepoCmd := &cobra.Command{
		Use:   "refresh [owner/name]",
		Short: "Refresh repository data",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			if len(args) == 0 {
				// Refresh all repositories
				err = client.RefreshAll()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error refreshing repositories: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("All repositories refreshed successfully")
			} else {
				// Refresh specific repository
				parts := strings.Split(args[0], "/")
				if len(parts) != 2 {
					fmt.Fprintf(os.Stderr, "Invalid repository name format, expected 'owner/repo'\n")
					os.Exit(1)
				}
				owner, name := parts[0], parts[1]

				err = client.RefreshRepository(owner, name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error refreshing repository: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Repository %s refreshed successfully\n", args[0])
			}
		},
	}

	// Pull request command
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Long:  "List and filter pull requests from tracked repositories",
	}

	// List pull requests command
	listPRCmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			// Get filter parameters
			params := make(map[string]string)
			params["state"], _ = cmd.Flags().GetString("state")
			params["author"], _ = cmd.Flags().GetString("author")
			params["repo"], _ = cmd.Flags().GetString("repo")
			params["sort"], _ = cmd.Flags().GetString("sort")
			params["direction"], _ = cmd.Flags().GetString("direction")
			page, _ := cmd.Flags().GetInt("page")
			perPage, _ := cmd.Flags().GetInt("per-page")
			params["page"] = fmt.Sprintf("%d", page)
			params["per_page"] = fmt.Sprintf("%d", perPage)

			resp, err := client.ListPullRequests(params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing pull requests: %v\n", err)
				os.Exit(1)
			}

			// Print pull requests
			fmt.Printf("%-40s %-5s %-20s %-12s %s\n", "REPOSITORY", "NUM", "AUTHOR", "STATE", "TITLE")
			for _, pr := range resp.Data {
				fmt.Printf("%-40s %-5d %-20s %-12s %s\n", pr.RepositoryFullName, pr.Number, pr.UserLogin, pr.State, pr.Title)
			}

			// Print pagination info
			fmt.Printf("\nPage %d of %d (Total: %d)\n", resp.Pagination.Page, resp.Pagination.TotalPages, resp.Pagination.Total)
		},
	}
	listPRCmd.Flags().StringP("state", "s", "open", "Filter by state (open, closed, all)")
	listPRCmd.Flags().StringP("author", "a", "", "Filter by author")
	listPRCmd.Flags().StringP("repo", "r", "", "Filter by repository (owner/name)")
	listPRCmd.Flags().String("sort", "created", "Sort by (created, updated)")
	listPRCmd.Flags().String("direction", "desc", "Sort direction (asc, desc)")
	listPRCmd.Flags().IntP("page", "p", 1, "Page number")
	listPRCmd.Flags().IntP("per-page", "n", 10, "Items per page")

	// Issue command
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issues",
		Long:  "List and filter issues from tracked repositories",
	}

	// List issues command
	listIssueCmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			// Get filter parameters
			params := make(map[string]string)
			params["state"], _ = cmd.Flags().GetString("state")
			params["author"], _ = cmd.Flags().GetString("author")
			params["repo"], _ = cmd.Flags().GetString("repo")
			params["sort"], _ = cmd.Flags().GetString("sort")
			params["direction"], _ = cmd.Flags().GetString("direction")
			page, _ := cmd.Flags().GetInt("page")
			perPage, _ := cmd.Flags().GetInt("per-page")
			params["page"] = fmt.Sprintf("%d", page)
			params["per_page"] = fmt.Sprintf("%d", perPage)

			resp, err := client.ListIssues(params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing issues: %v\n", err)
				os.Exit(1)
			}

			// Print issues
			fmt.Printf("%-40s %-5s %-20s %-12s %s\n", "REPOSITORY", "NUM", "AUTHOR", "STATE", "TITLE")
			for _, issue := range resp.Data {
				fmt.Printf("%-40s %-5d %-20s %-12s %s\n", issue.RepositoryFullName, issue.Number, issue.UserLogin, issue.State, issue.Title)
			}

			// Print pagination info
			fmt.Printf("\nPage %d of %d (Total: %d)\n", resp.Pagination.Page, resp.Pagination.TotalPages, resp.Pagination.Total)
		},
	}
	listIssueCmd.Flags().StringP("state", "s", "open", "Filter by state (open, closed, all)")
	listIssueCmd.Flags().StringP("author", "a", "", "Filter by author")
	listIssueCmd.Flags().StringP("repo", "r", "", "Filter by repository (owner/name)")
	listIssueCmd.Flags().String("sort", "created", "Sort by (created, updated)")
	listIssueCmd.Flags().String("direction", "desc", "Sort direction (asc, desc)")
	listIssueCmd.Flags().IntP("page", "p", 1, "Page number")
	listIssueCmd.Flags().IntP("per-page", "n", 10, "Items per page")

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := NewClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
				os.Exit(1)
			}

			status, err := client.GetStatus()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
				os.Exit(1)
			}

			// Print status
			fmt.Println("Service Status:")
			fmt.Printf("  Status: %s\n", status["status"])
			fmt.Printf("  Version: %s\n", status["version"])

			// Print repository stats
			if repoStats, ok := status["repositories"].(map[string]interface{}); ok {
				fmt.Println("\nRepositories:")
				fmt.Printf("  Total: %v\n", repoStats["total"])
				fmt.Printf("  Syncing: %v\n", repoStats["syncing"])
				fmt.Printf("  Error: %v\n", repoStats["error"])
			}

			// Print GitHub rate limit
			if rateLimit, ok := status["github_rate_limit"].(map[string]interface{}); ok {
				fmt.Println("\nGitHub Rate Limit:")
				fmt.Printf("  Limit: %v\n", rateLimit["limit"])
				fmt.Printf("  Remaining: %v\n", rateLimit["remaining"])
				if resetAt, ok := rateLimit["reset_at"].(string); ok {
					fmt.Printf("  Reset At: %s\n", resetAt)
				}
			}
		},
	}

	// Add commands to repo command
	repoCmd.AddCommand(addRepoCmd, listRepoCmd, removeRepoCmd, refreshRepoCmd)

	// Add commands to pr command
	prCmd.AddCommand(listPRCmd)

	// Add commands to issue command
	issueCmd.AddCommand(listIssueCmd)

	// Add commands to root command
	rootCmd.AddCommand(repoCmd, prCmd, issueCmd, statusCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
