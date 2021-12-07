# Actions Runner Manager

[![Continuous Deployment](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml/badge.svg)](https://github.com/lindluni/actions-runner-manager/actions/workflows/merge.yml)

Actions Runner Manager is a GitHub Application that can be used by users who are not organization owners to manage
GitHub Actions Organization Runner Groups. Actions Runner Manager implements pseudo-RBAC policies built on top of GitHub 
Teams and exposes a set of authenticated API's to grant access to the GitHub Organization Self-Hosted Runner REST API's
which are generally only available to organization owners.

**Notice**: Actions Runner Manager does not currently support GitHub Enterprise Server

## Authorization

Actions Runner Manager uses existing GitHub Teams to create a pseudo-RBAC policy. Every call requires a user to submit a valid GitHub API token
assigned to a user who is a maintainer of a GitHub Team in the `Authorization` header.

**Notice**: Only users who are maintainers of a GitHub Team may use the Actions Runner Manager API's.

When a user makes a request to any of the Actions Runner Manager API's, the `team` parameter is used as the name of the
Runner Group. When a user calls `/group-create` or `/group-delete`, an Organization Runner Group is created or deleted
with the name specified in the `team` parameter. All other API endpoints query or delete assets from the Organization 
Runner Group assigned to the `team` parameter.

When a user submits a valid GitHub API token to the Actions Runner Manager API's, the application first creates an
authorized GitHub client using the users GitHub API token. The API then makes an authenticated API call as the
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
tightly scoped to only the Teams they need access to in order to limit risk and exposure.


## Rate Limiting

To protect the integrity of the server, Actions Runner Manager uses a rate limit cache to enforce an admin configured
rate limit policy. The rate limit cache is purged every 60 minutes. The rate limit setting defines the maximum number of
authenticated requests a user can make per second. If you encounter a rate limit error, you should wait a few seconds
and attempt your request again.

## GitHub Application Configuration

To configure the Actions Runner Manager, you must create and install a GitHub App. The app must be configured with
the following permissions:

```text
Repository:
    Metadata: Read-Only
Organization:
    Members: Read-Only
    Administration: Read and Write
    Self-Hosted Runners: Read and Write
```

**Note**: The Actions Runner Manager does not currently make use of Webhooks or Callbacks, you may leave these options
blank when creating the GitHub Application. 

Once you have created a GitHub App, you must install the application in your organization and give it permission
to the repositories you want it to be able to assign runner groups to.

You can follow GitHub's documentation on how to create and install a GitHub App here:
- [Create a GitHub App](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app)
- [Installing a GitHub App](https://docs.github.com/en/developers/apps/managing-github-apps/installing-github-apps)

## Actions Runner Manager Configuration

The Actions Manager reads a static YAML file to create its configuration. By default, the config file is read from the
directory in which the application is run from with a file name of `config.yml`. Users can override the default path by
setting the `CONFIG_PATH` environment variable, for example: `export CONFIG_PATH=/opt/arm/config.yml`.

The configuration file is a YAML file with the following structure:
```yaml
org: "<GitHub Organization>"
appID: <GitHub Application ID>
installationID: <GitHub Application Installation ID>
privateKey: "<Base64 Encoded GitHub Application Private Key>"
rateLimit: <Maximum number of authenticated requests a user can make per second>
logging:
  compress: (true or false) <Compress rotated log files>
  ephemeral: (true or false) <Log to stdout instead of rotating log files>
  level: (debug, info, warn, error, or fatal) <Logging level>
  logDirectory: <Relative or absolut directory to log to>
  maxAge: <Maximum number of days to keep log files>
  maxBackups: <Maximum number of log files to keep>
  maxSize: <Maximum size of log files in bytes before rotation>
server:
  address: "<IP Address or Hostname bind interface>"
  port: <Port to bind to>
  tls:
    enabled: (true or false) <Enable TLS>
    certFile: "<Path to TLS certificate file>"
    keyFile: "<Path to TLS key file>"
```

**Note**: You can encode your private into Base64 by using the following command after downloading it from the GitHub UI:

```shell
    cat <private_key_file> | base64
```

**Warning**: Because users are expected to pass their Authorization tokens in the Authorization header, you should never
run the Actions Runner Manager in production with TLS disabled.

## Running the Server

**Security Notice**: Actions Runner Manager should never run in non-TLS mode when in production. Users should configure
TLS on the server directly or use some form of software or hardware to enforce TLS.

---

### Standalone

Download the current binary from: https://github.com/lindluni/actions-runner-manager/releases

or

Build the current binary using the Go toolchain: 
```shell
    go install github.com/lindluni/actions-runner-manager/pkg
```

Then create a config file according to the documentation above, then run the binary with the following command:
```shell
    actions-runner-manager
```

It is recommended you use a Service Manager such as systemd to ensure the server is running.

---

### Docker

Actions Runner Manager hosts its image on the GitHub Container Registry. You must first authenticate using a GitHub
API token to pull the image using the following command:

```shell
    docker login -u <GitHub Username> -p <GitHub API Token> ghcr.io
```

Create a config file according to the documentation above, then run the following command:

```shell
    docker run -it -d --restart always \
    -v <absolute_path_to_config_file>:<config.yml> \
    -p <local port>:<port set in config> \
    ghcr.io/lindluni/actions-runner-manager:latest
```

---

### Kubernetes Using Helm

**TO BE COMPLETED**

---

## API's

You can find detailed API documentation on our [GitHub Page](https://lindluni.github.io/actions-runner-manager/)

Actions Runner Manager implements the following API's:

---

#### `/api/v1/group-add`

- Create a new GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-add?team=<team_slug>"
```

---

#### `/api/v1/group-delete`

- Delete an existing GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-delete?team=<team_slug>"
```

---

#### `/api/v1/group-list`

- List all the runners and repositories assigned to a GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/group-list?team=<team_slug>"
```

---

#### `/api/v1/repos-add`

- Add one or more repositories to an existing GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-add?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

---

#### `/api/v1/repos-remove`

- Remove one or more repositories from an existing GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-remove?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

---

#### `/api/v1/repos-set`

- Replace all existing repositories assigned to an existing GitHub Actions Runner Group with the name in the `team` parameter with one or more new repositories

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/repos-set?team=<team_slug>&repos=<repo1>,<repo2>,<repo3>"
```

---

#### `/api/v1/token-register`

- Create a new Registration Token to be used during runner configuration to register a runner to an existing GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/token-register?team=<team_slug>"
```

---

#### `/api/v1/token-remove`

- Create a new Removal Token to be used during runner de-provisioning to remove a runner from an existing GitHub Actions Organization Runner Group with the name in the `team` parameter

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/token-remove?team=<team_slug>"
```

---

#### `/api/v1/status`

- Checks the readiness status of the server

```shell
curl -H "Authorization: <token>" "https://<host>:<port>/api/v1/status"
```

## Why Distroless?

The Google Distroless containers provide a simple, secure, and scalable way to run Docker containers. The Distroless image
is a small base image and contains no executables or shell other than the Actions Runner Manager and its dependencies.
As such, it is ultra secure as it contains no extraneous dependencies requiring being kept up to date and potentially exposing
the application API's to outside vulnerabilities.

The major caveat of the Distroless container is that it contains no shell, thus making it more secure, but as such, it cannot 
be exec'ed into. If you require a shell, you must modify the Dockerfile to use a different base image for the final container.

You can read up on the Google Distroless base image at the links below:

- [Google Documentation](https://github.com/GoogleContainerTools/distroless)
- [Why Kubernetes Switched to Distroless](https://github.com/kubernetes/enhancements/blob/master/keps/sig-release/1729-rebase-images-to-distroless/README.md)
