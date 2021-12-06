## Actions Runner Manager

[![Merge Tests](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml/badge.svg)](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml)

Actions Runner Manager is a GitHub Application that can be used by users who are not organization owners to manage
GitHub Actions Organization Runner Groups. Actions Runner Manager implements RBAC policies built on top of GitHub 
Teams and exposes a set of authenticated API's to grant access to the GitHub Organization Self-Hosted Runner REST API's.

### Authorization

Actions Runner Manager uses existing GitHub Teams to manage RBAC policies. Every call requires a user to submit a valid GitHub API token
assigned to a user who is a maintainer of a GitHub Team in the `Authorization` header.

**Notice**: Only users who are maintainers of a GitHub Team may use the Actions Runner Manager API's.

When a user submits a valid GitHub API token to the Actions Runner Manager API's, the application first creates an
authorized GitHub client using the users GitHub API token. The API then makes a single authenticated API call as the
user to the `/user` GitHub API endpoint documented here: https://docs.github.com/en/rest/reference/users#get-the-authenticated-user

This call returns the GitHub Users object of which the only information retrieved is the users `login` property. This
property is the GitHub username of the authenticated user. This data is passed to the Teams API to confirm the user
making the request on the Actions Runner Manager API is a maintainer of the GitHub Team. The users API token is not used
for any other purpose. The username is also stored in the rate limit cache for 60 minutes to enforce our rate limit
policies on the server. If the user has not made an authenticated API call in the past 60 minutes, the rate limit cache
purges the username of the authenticated user.

**Note**: While the Actions Runner Manager API's make secure, limited use of the users object, and does not call any
other API endpoints while authenticated as the user, users should be sensitive to the fact that the Users API returns
private Personally Identifiable Information (PII) such as email addresses. As such, we recommend users use bot accounts
tightly scoped to only the Teams and Repos they need access to in order to limit risk and exposure.


### Rate Limiting

To protect the integrity of the server, Actions Runner Manager uses a rate limit cache to enforce an admin configured
rate limit policy. The rate limit cache is purged every 60 minutes. The rate limit setting defines the maximum number of
authenticated requests a user can make per second. If you encounter a rate limit error, you should wait a few seconds
and attempt your request again.

### Sample Configuration

### API's

You can find detailed API documentation on our [GitHub Page](https://lindluni.github.io/actions-runner-manager/)

Actions Runner Manager implements the following API's:

### `/api/v1/group-add`

- Create a new GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-add?team=<team_slug>"
```

### `/api/v1/group-delete`

- Delete an existing GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-delete?team=<team_slug>"
```

### `/api/v1/group-list`

- List all the runners and repositories assigned to a GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-list?team=<team_slug>"
```

### `/api/v1/repos-add`

- Add one or more repositories to an existing GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-add?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

### `/api/v1/repos-remove`

- Remove one or more repositories from an existing GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-remove?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

### `/api/v1/repos-set`

- Replace all of the existing repositories assigned to an existing GitHub Actions Runner Group with one or more new repositories

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-set?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

### `/api/v1/token-register`

- Create a new Registration Token to be used during runner configuration to register a runner to an existing GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/token-register?team=<team_slug>"
```

### `/api/v1/token-remove`

- Create a new Removal Token to be used during runner de-provisioning to remove a runner from an existing GitHub Actions Organization Runner Group

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/token-remove?team=<team_slug>"
```

## Why Distroless?

https://github.com/GoogleContainerTools/distroless
https://github.com/kubernetes/enhancements/blob/master/keps/sig-release/1729-rebase-images-to-distroless/README.md
