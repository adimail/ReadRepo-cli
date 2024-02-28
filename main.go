package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

type RepoInfo struct {
	Name               string    `json:"name"`
	Owner              Owner     `json:"owner"`
	Description        string    `json:"description"`
	License            License   `json:"license"`
	CreatedAt          time.Time `json:"created_at"`
	DaysSinceCreation  int
	StargazersCount    int    `json:"stargazers_count"`
	ForksCount         int    `json:"forks_count"`
	Branches           int
	CommitsCount       int
	FilesCount         int
	LanguagesUsed      []string
	ReleasesCount      int
	WorkflowsCount     int
	IssuesCount        int
	PullsCount         int
	ContributorsCount  int
	LastCommitDate     time.Time
	CodeFrequencyLink  string
}

type Owner struct {
	Login string `json:"login"`
}

type License struct {
	Name string `json:"name"`
}

func getGitHubRepoInfo(repoURL string, token string) (*RepoInfo, error) {
	repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")

	// Ensure we have at least two parts
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository URL format")
	}

	// Extract owner and repo
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repository information, status code: %d", resp.StatusCode)
	}

	var repoInfo RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}

	// Get number of branches
	branchesResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo))
	if err != nil {
		return nil, err
	}
	defer branchesResp.Body.Close()
	branchesCount := len(getJSONData(branchesResp))

	// Get number of contributors
	contributorsResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors", owner, repo))
	if err != nil {
		return nil, err
	}
	defer contributorsResp.Body.Close()
	contributorsCount := len(getJSONData(contributorsResp))

	// Get number of total commits
	commitsResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/commits", owner, repo))
	if err != nil {
		return nil, err
	}
	defer commitsResp.Body.Close()
	commitsCount := len(getJSONData(commitsResp))

	// Get number of total files
	filesResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", owner, repo))
	if err != nil {
		return nil, err
	}
	defer filesResp.Body.Close()
	filesCount := len(getJSONData(filesResp))

	// Get languages used
	languagesResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", owner, repo))
	if err != nil {
		return nil, err
	}
	defer languagesResp.Body.Close()
	var languagesMap map[string]interface{}
	if err := json.NewDecoder(languagesResp.Body).Decode(&languagesMap); err != nil {
		return nil, err
	}
	var languagesUsed []string
	for lang := range languagesMap {
		languagesUsed = append(languagesUsed, lang)
	}

	// Get releases
	releasesResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo))
	if err != nil {
		return nil, err
	}
	defer releasesResp.Body.Close()
	releasesCount := len(getJSONData(releasesResp))

	// Get workflows (actions)
	workflowsResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows", owner, repo))
	if err != nil {
		return nil, err
	}
	defer workflowsResp.Body.Close()
	var workflowsData map[string]interface{}
	if err := json.NewDecoder(workflowsResp.Body).Decode(&workflowsData); err != nil {
		return nil, err
	}
	workflowsCount := len(workflowsData["workflows"].([]interface{}))

	// Get number of issues
	issuesResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo))
	if err != nil {
		return nil, err
	}
	defer issuesResp.Body.Close()
	issuesCount := len(getJSONData(issuesResp))

	// Get number of open pull requests
	pullsResp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo))
	if err != nil {
		return nil, err
	}
	defer pullsResp.Body.Close()
	pullsCount := len(getJSONData(pullsResp))

	// Get code frequency link
	codeFrequencyLink := fmt.Sprintf("https://github.com/%s/%s/graphs/code-frequency", owner, repo)

	// Get the date of the last commit
	lastCommitDate, err := getLastCommitDate(owner, repo, token)
	if err != nil {
		return nil, err
	}

	// Calculate days since creation
	daysSinceCreation := int(time.Since(repoInfo.CreatedAt).Hours() / 24)

	// Add additional information to the repo info
	repoInfo.Branches = branchesCount
	repoInfo.CommitsCount = commitsCount
	repoInfo.FilesCount = filesCount
	repoInfo.LanguagesUsed = languagesUsed
	repoInfo.ReleasesCount = releasesCount
	repoInfo.WorkflowsCount = workflowsCount
	repoInfo.IssuesCount = issuesCount
	repoInfo.PullsCount = pullsCount
	repoInfo.ContributorsCount = contributorsCount
	repoInfo.LastCommitDate = lastCommitDate
	repoInfo.DaysSinceCreation = daysSinceCreation
	repoInfo.CodeFrequencyLink = codeFrequencyLink

	return &repoInfo, nil
}

func getLastCommitDate(owner, repo, token string) (time.Time, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/commits", owner, repo), nil)
	if err != nil {
		return time.Time{}, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("failed to fetch commits, status code: %d", resp.StatusCode)
	}

	var commits []struct {
		Commit struct {
			Author struct {
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return time.Time{}, err
	}

	if len(commits) > 0 {
		lastCommitDate, err := time.Parse(time.RFC3339, commits[0].Commit.Author.Date)
		if err != nil {
			return time.Time{}, err
		}
		return lastCommitDate, nil
	}

	return time.Time{}, fmt.Errorf("no commits found")
}

func getJSONData(resp *http.Response) []interface{} {
	var data []interface{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return data
	}
	return data
}

func printRepoInfo(repoInfo *RepoInfo) {
	fmt.Println("Repository Name:", repoInfo.Name)
	fmt.Println("Owner:", repoInfo.Owner.Login)
	fmt.Println("Description:", repoInfo.Description)
	fmt.Println("License:", repoInfo.License.Name)
	fmt.Println("Date Created:", repoInfo.CreatedAt)
	fmt.Println("Days Since Creation:", repoInfo.DaysSinceCreation)
	fmt.Println("Stars:", repoInfo.StargazersCount)
	fmt.Println("Forks:", repoInfo.ForksCount)
	fmt.Println("Branches:", repoInfo.Branches)
	fmt.Println("Number of Commits:", repoInfo.CommitsCount)
	fmt.Println("Total Number of Files:", repoInfo.FilesCount)
	fmt.Println("Languages Used:", strings.Join(repoInfo.LanguagesUsed, ", "))
	fmt.Println("Releases:", repoInfo.ReleasesCount)
	fmt.Println("Workflows (Actions):", repoInfo.WorkflowsCount)
	fmt.Println("Issues:", repoInfo.IssuesCount)
	fmt.Println("Pull Requests:", repoInfo.PullsCount)
	fmt.Println("Total Number of Contributors:", repoInfo.ContributorsCount)
	fmt.Println("Last Commit Date:", repoInfo.LastCommitDate)
	fmt.Println("Code Frequency:", repoInfo.CodeFrequencyLink)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <github_repo_url> <github_token>")
		return
	}

	githubRepoURL := os.Args[1]
	githubToken := os.Args[2]

	// Create a new spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Prefix = "Calculating... "
	s.Start()

	// Fetch repository info in a separate goroutine
	var repoInfo *RepoInfo
	var err error
	go func() {
		repoInfo, err = getGitHubRepoInfo(githubRepoURL, githubToken)
	}()

	// Wait for the fetch to complete
	for repoInfo == nil && err == nil {
		time.Sleep(100 * time.Millisecond)
	}

	// Stop the spinner
	s.Stop()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	printRepoInfo(repoInfo)
}