package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL  string
	client  *Client
	page    int
	perPage int
	format  string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "ghrepos",
	Short: "GitHub Repository Management CLI",
	Long: `A command-line interface for the GitHub Repository Management Service.
This CLI allows you to manage GitHub repositories, pull requests, and issues.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		client = NewClient(apiURL)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().IntVar(&page, "page", 1, "Page number for pagination")
	rootCmd.PersistentFlags().IntVar(&perPage, "per-page", 30, "Number of items per page")
	rootCmd.PersistentFlags().StringVar(&format, "format", "table", "Output format (table, json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")

	// Add repository commands
	rootCmd.AddCommand(newRepoCmd())

	// Add pull request commands
	rootCmd.AddCommand(newPRCmd())

	// Add issue commands
	rootCmd.AddCommand(newIssueCmd())

	// Add service commands
	rootCmd.AddCommand(newServiceCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Repository commands
func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
		Long:  `Add, remove, and list GitHub repositories being tracked by the service.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all tracked repositories",
			Run: func(cmd *cobra.Command, args []string) {
				// Get repositories
				resp, err := client.ListRepositories(page, perPage)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				// Print repositories
				if format == "json" {
					printJSON(resp)
					return
				}

				// Print table
				fmt.Printf("Repositories (%d/%d):\n", len(resp.Data), resp.Pagination.Total)
				fmt.Println("OWNER/NAME\tPRIVATE\tLAST SYNCED")
				fmt.Println("----------\t-------\t-----------")
				for _, repo := range resp.Data {
					lastSynced := "Never"
					if repo.LastSyncedAt != "" && repo.LastSyncedAt != "0001-01-01T00:00:00Z" {
						parsedTime, err := time.Parse(time.RFC3339, repo.LastSyncedAt)
						if err == nil {
							lastSynced = parsedTime.Format("2006-01-02 15:04:05")
						}
					}
					fmt.Printf("%s/%s\t%v\t%s\n", repo.Owner, repo.Name, repo.IsPrivate, lastSynced)
				}
				fmt.Printf("\nPage %d of %d\n", resp.Pagination.Page, resp.Pagination.TotalPages)
			},
		},
		&cobra.Command{
			Use:   "add [owner/repo]",
			Short: "Add a repository to track",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Add repository
				repo, err := client.AddRepository(args[0])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				// Print repository
				if format == "json" {
					printJSON(repo)
					return
				}

				fmt.Printf("Repository added: %s/%s\n", repo.Owner, repo.Name)
			},
		},
		&cobra.Command{
			Use:   "remove [owner/repo]",
			Short: "Remove a repository from tracking",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Parse owner and repo
				parts := strings.Split(args[0], "/")
				if len(parts) != 2 {
					fmt.Println("Error: Invalid repository name format, expected 'owner/repo'")
					return
				}
				owner, repo := parts[0], parts[1]

				// Remove repository
				err := client.RemoveRepository(owner, repo)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				fmt.Printf("Repository removed: %s/%s\n", owner, repo)
			},
		},
		&cobra.Command{
			Use:   "refresh [owner/repo]",
			Short: "Force refresh of repository data",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Parse owner and repo
				parts := strings.Split(args[0], "/")
				if len(parts) != 2 {
					fmt.Println("Error: Invalid repository name format, expected 'owner/repo'")
					return
				}
				owner, repo := parts[0], parts[1]

				// Refresh repository
				err := client.RefreshRepository(owner, repo)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				fmt.Printf("Repository refresh initiated: %s/%s\n", owner, repo)
			},
		},
	)

	return cmd
}

// Pull request commands
func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Long:  `List and filter pull requests across all tracked repositories.`,
	}

	// Add subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		Run: func(cmd *cobra.Command, args []string) {
			// Parse flags
			state, _ := cmd.Flags().GetString("state")
			author, _ := cmd.Flags().GetString("author")
			repo, _ := cmd.Flags().GetString("repo")
			label, _ := cmd.Flags().GetString("label")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			direction, _ := cmd.Flags().GetString("direction")
			since, _ := cmd.Flags().GetString("since")
			groupBy, _ := cmd.Flags().GetString("group-by")

			// Build params
			params := make(map[string]string)
			params["page"] = strconv.Itoa(page)
			params["per_page"] = strconv.Itoa(perPage)
			if state != "" {
				params["state"] = state
			}
			if author != "" {
				params["author"] = author
			}
			if repo != "" {
				params["repo"] = repo
			}
			if label != "" {
				params["label"] = label
			}
			if sortBy != "" {
				params["sort"] = sortBy
			}
			if direction != "" {
				params["direction"] = direction
			}
			if since != "" {
				params["since"] = since
			}
			if groupBy != "" {
				params["group_by"] = groupBy
			}

			// Get pull requests
			resp, err := client.ListPullRequests(params)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Print pull requests
			if format == "json" {
				printJSON(resp)
				return
			}

			// Print table
			fmt.Printf("Pull Requests (%d/%d):\n", len(resp.Data), resp.Pagination.Total)
			fmt.Println("NUMBER\tTITLE\tSTATE\tAUTHOR\tUPDATED")
			fmt.Println("------\t-----\t-----\t------\t-------")
			for _, pr := range resp.Data {
				updatedAt := "Never"
				if pr.UpdatedAt != "" && pr.UpdatedAt != "0001-01-01T00:00:00Z" {
					parsedTime, err := time.Parse(time.RFC3339, pr.UpdatedAt)
					if err == nil {
						updatedAt = parsedTime.Format("2006-01-02 15:04:05")
					}
				}
				fmt.Printf("#%d\t%s\t%s\t%s\t%s\n", pr.Number, truncate(pr.Title, 30), pr.State, pr.UserLogin, updatedAt)
			}
			fmt.Printf("\nPage %d of %d\n", resp.Pagination.Page, resp.Pagination.TotalPages)
		},
	}

	// Add flags for filtering
	listCmd.Flags().String("state", "open", "Filter by PR state (open, closed, all)")
	listCmd.Flags().String("author", "", "Filter by PR author")
	listCmd.Flags().String("repo", "", "Filter by repository (owner/repo)")
	listCmd.Flags().String("label", "", "Filter by label")
	listCmd.Flags().String("sort-by", "", "Sort by field")
	listCmd.Flags().String("direction", "", "Sort direction")
	listCmd.Flags().String("since", "", "Filter pull requests since a certain date")
	listCmd.Flags().String("group-by", "", "Group pull requests by field")

	cmd.AddCommand(listCmd)

	return cmd
}

// Issue commands
func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issues",
		Long:  `List and filter issues across all tracked repositories.`,
	}

	// Add subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		Run: func(cmd *cobra.Command, args []string) {
			// Parse flags
			state, _ := cmd.Flags().GetString("state")
			author, _ := cmd.Flags().GetString("author")
			repo, _ := cmd.Flags().GetString("repo")
			label, _ := cmd.Flags().GetString("label")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			direction, _ := cmd.Flags().GetString("direction")
			since, _ := cmd.Flags().GetString("since")
			groupBy, _ := cmd.Flags().GetString("group-by")

			// Build params
			params := make(map[string]string)
			params["page"] = strconv.Itoa(page)
			params["per_page"] = strconv.Itoa(perPage)
			if state != "" {
				params["state"] = state
			}
			if author != "" {
				params["author"] = author
			}
			if repo != "" {
				params["repo"] = repo
			}
			if label != "" {
				params["label"] = label
			}
			if sortBy != "" {
				params["sort"] = sortBy
			}
			if direction != "" {
				params["direction"] = direction
			}
			if since != "" {
				params["since"] = since
			}
			if groupBy != "" {
				params["group_by"] = groupBy
			}

			// Get issues
			resp, err := client.ListIssues(params)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Print issues
			if format == "json" {
				printJSON(resp)
				return
			}

			// Print table
			fmt.Printf("Issues (%d/%d):\n", len(resp.Data), resp.Pagination.Total)
			fmt.Println("NUMBER\tTITLE\tSTATE\tAUTHOR\tUPDATED")
			fmt.Println("------\t-----\t-----\t------\t-------")
			for _, issue := range resp.Data {
				updatedAt := "Never"
				if issue.UpdatedAt != "" && issue.UpdatedAt != "0001-01-01T00:00:00Z" {
					parsedTime, err := time.Parse(time.RFC3339, issue.UpdatedAt)
					if err == nil {
						updatedAt = parsedTime.Format("2006-01-02 15:04:05")
					}
				}
				fmt.Printf("#%d\t%s\t%s\t%s\t%s\n", issue.Number, truncate(issue.Title, 30), issue.State, issue.UserLogin, updatedAt)
			}
			fmt.Printf("\nPage %d of %d\n", resp.Pagination.Page, resp.Pagination.TotalPages)
		},
	}

	// Add flags for filtering
	listCmd.Flags().String("state", "open", "Filter by issue state (open, closed, all)")
	listCmd.Flags().String("author", "", "Filter by issue author")
	listCmd.Flags().String("repo", "", "Filter by repository (owner/repo)")
	listCmd.Flags().String("label", "", "Filter by label")
	listCmd.Flags().String("sort-by", "", "Sort by field")
	listCmd.Flags().String("direction", "", "Sort direction")
	listCmd.Flags().String("since", "", "Filter issues since a certain date")
	listCmd.Flags().String("group-by", "", "Group issues by field")

	cmd.AddCommand(listCmd)

	return cmd
}

// Service commands
func newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage the service",
		Long:  `Manage the GitHub Repository Management Service.`,
	}

	// Add subcommands
	cmd.AddCommand(
		&cobra.Command{
			Use:   "refresh",
			Short: "Force refresh of GitHub data",
			Run: func(cmd *cobra.Command, args []string) {
				// Refresh all data
				err := client.RefreshAll()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				fmt.Println("Refresh initiated for all repositories")
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Get service status",
			Run: func(cmd *cobra.Command, args []string) {
				// Get status
				status, err := client.GetStatus()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}

				// Print status
				if format == "json" {
					printJSON(status)
					return
				}

				// Print table
				fmt.Printf("Service Status: %s\n", status["status"])
				fmt.Printf("Version: %s\n", status["version"])
				fmt.Printf("Uptime: %d seconds\n", int(status["uptime"].(float64)))

				// Print repositories
				repos := status["repositories"].(map[string]interface{})
				fmt.Printf("Repositories: %d total, %d syncing, %d error\n", int(repos["total"].(float64)), int(repos["syncing"].(float64)), int(repos["error"].(float64)))

				// Print rate limit
				rateLimit := status["github_rate_limit"].(map[string]interface{})
				resetAt, _ := time.Parse(time.RFC3339, rateLimit["reset_at"].(string))
				fmt.Printf("GitHub Rate Limit: %d/%d (resets at %s)\n", int(rateLimit["remaining"].(float64)), int(rateLimit["limit"].(float64)), resetAt.Format("15:04:05"))
			},
		},
	)

	return cmd
}

// Helper functions

// printJSON prints data as JSON
func printJSON(data interface{}) {
	// TODO: Implement JSON output
	fmt.Printf("%+v\n", data)
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
