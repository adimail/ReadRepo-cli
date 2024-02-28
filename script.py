import requests
from datetime import datetime, timezone

def get_github_api_response(url, headers=None):
    with requests.Session() as session:
        try:
            response = session.get(url, headers=headers)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Error fetching data from {url}: {e}")
            return None

def get_repo_info(repo_url, token=None):
    if repo_url.endswith(".git"):
        repo_url = repo_url[:-4]
    parts = repo_url.strip("/").split("/")
    owner, repo = parts[-2:]

    headers = {'Authorization': f'token {token}'} if token else {}

    repo_info_url = f"https://api.github.com/repos/{owner}/{repo}"
    repo_info = get_github_api_response(repo_info_url, headers)

    if repo_info:
        repo_info['owner'] = owner
        repo_info['branches'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/branches", headers))
        repo_info['contributors_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/contributors", headers))
        repo_info['commits_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/commits", headers))
        repo_info['files_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/contents", headers))
        repo_info['languages_used'] = list(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/languages", headers).keys())
        repo_info['releases_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/releases", headers))
        repo_info['workflows_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/actions/workflows", headers)["workflows"])
        repo_info['issues_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/issues", headers))
        repo_info['pulls_count'] = len(get_github_api_response(f"https://api.github.com/repos/{owner}/{repo}/pulls", headers))

        repo_info['code_frequency_link'] = f"https://github.com/{owner}/{repo}/graphs/code-frequency"

        created_at = datetime.strptime(repo_info['created_at'], '%Y-%m-%dT%H:%M:%SZ').replace(tzinfo=timezone.utc)
        repo_info['days_since_creation'] = (datetime.now(timezone.utc) - created_at).days

        last_commit_date = datetime.strptime(repo_info['updated_at'], '%Y-%m-%dT%H:%M:%SZ').replace(tzinfo=timezone.utc)
        repo_info['last_commit_date'] = last_commit_date

        return repo_info
    else:
        return None

def print_repo_info(repo_info):
    print("Repository Name:", repo_info.get("name"))
    print("Owner:", repo_info.get("owner"))
    print("Description:", repo_info.get("description"))
    print("License:", repo_info.get("license", {}).get("name", "N/A"))
    print("Date Created:", repo_info.get("created_at"))
    print("Days Since Creation:", repo_info.get("days_since_creation"))
    print("Stars:", repo_info.get("stargazers_count"))
    print("Forks:", repo_info.get("forks_count"))
    print("Branches:", repo_info.get("branches"))
    print("Number of Commits:", repo_info.get("commits_count"))
    print("Total Number of Files:", repo_info.get("files_count"))
    print("Languages Used:", ', '.join(repo_info.get("languages_used", [])))
    print("Releases:", repo_info.get("releases_count"))
    print("Workflows (Actions):", repo_info.get("workflows_count"))
    print("Issues:", repo_info.get("issues_count"))
    print("Pull Requests:", repo_info.get("pulls_count"))
    print("Total Number of Contributors:", repo_info.get("contributors_count"))
    print("Last Commit Date:", repo_info.get("last_commit_date"))
    print("Code Frequency:", repo_info.get("code_frequency_link"))

if __name__ == "__main__":
    github_repo_url = input("Enter GitHub repository URL: ")
    github_token = input("Enter your GitHub personal access token (optional): ")
    repo_info = get_repo_info(github_repo_url, github_token)

    if repo_info:
        print_repo_info(repo_info)
