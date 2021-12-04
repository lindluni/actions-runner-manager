## Actions Runner Manager

[![Merge Tests](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml/badge.svg)](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml)

Actions Runner Manager is a GitHub Application that can be used by users who are not organization owners to manage
GitHub Actions Organization Runner Groups. Actions Runner Manager implements RBAC policies built on top of GitHub 
Teams and exposes a set of authenticated API's to grant access to the GitHub Organization Self-Hosted Runner REST API's.

### API's

Actions Runner Manager implements the following API's:

---

```shell
[GET /v1/api/group-create]
```

- Creates a new GitHub Actions Organization Runner Group with a slug matching the team slug submitted in the `team` parameter 
```text
Headers:
-----------
Authorization: (Required) -- A GitHub Personal Access token belonging to a maintainer of the team parameter
-----------

Parameters:
-----------
team: (Required) -- The canonical team slug of a GitHub Team that the Authenticated user is a maintainer of
-----------
```

Status Codes:
```shell
200 -- Success
401 -- GitHub Personal Access token is missing or invalid
403 -- GitHub Team does not exist
409 -- GitHub Actions runner group already exists
```

Example:
```shell
curl -H "Authorization: token gh_test-token" https:<host>/group-create?team=test-team
```

Example Response:
```shell
{"Response":"Runner group created successfully: test-team","StatusCode":200}
```

---

### `[GET /v1/api/group-delete]`

---

### `[GET /v1/api/group-list]`

---

### `[GET /v1/api/repos-add]`

---

### `[GET /v1/api/repos-remove]`

---

### `[GET /v1/api/repos-set]`

---

### `[GET /v1/api/token-register]`

---

###`[GET /v1/api/token-remove]`

