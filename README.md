# Factorio watchdog for release changes

Build trigger for factorio on new releases

## Enviroment variables

Factorio Docker Watchdog must be configured via environment variables.
Below you find a list with all required and optional variables.

### Required

- **GIT_EMAIL**: Email used by git for committing
- **GIT_NAME**: Username used by git for committing
- **GITHUB_USER**: GitHub Username used for api calls, git pull and push
- **GITHUB_TOKEN**: GitHub token used for api calls, git pull and push
- **GITHUB_REPO_OWNER**: GitHub Username/Organization where to push to
- **GITHUB_REPO_NAME**: GitHub Repository where to push to

### Optional

- **DISCORD_WEBHOOK_URL**: Discord webhook url for announcements
