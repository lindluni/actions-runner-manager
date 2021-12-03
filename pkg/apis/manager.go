package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	"github.com/sirupsen/logrus"
)

//go:generate counterfeiter -generate

//counterfeiter:generate -o mocks/actions_client.go -fake-name ActionsClient . actionsClient
type actionsClient interface {
	AddRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID, repoID int64) (*github.Response, error)
	CreateOrganizationRegistrationToken(ctx context.Context, owner string) (*github.RegistrationToken, *github.Response, error)
	CreateOrganizationRemoveToken(ctx context.Context, owner string) (*github.RemoveToken, *github.Response, error)
	CreateOrganizationRunnerGroup(ctx context.Context, org string, createReq github.CreateRunnerGroupRequest) (*github.RunnerGroup, *github.Response, error)
	DeleteOrganizationRunnerGroup(ctx context.Context, org string, groupID int64) (*github.Response, error)
	ListOrganizationRunnerGroups(ctx context.Context, org string, opts *github.ListOptions) (*github.RunnerGroups, *github.Response, error)
	ListRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID int64, opts *github.ListOptions) (*github.ListRepositories, *github.Response, error)
	ListRunnerGroupRunners(ctx context.Context, org string, groupID int64, opts *github.ListOptions) (*github.Runners, *github.Response, error)
	RemoveRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID, repoID int64) (*github.Response, error)
	SetRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID int64, ids github.SetRepoAccessRunnerGroupRequest) (*github.Response, error)
}

//counterfeiter:generate -o mocks/teams_client.go -fake-name TeamsClient . teamsClient
type teamsClient interface {
	GetTeamMembershipBySlug(ctx context.Context, org, slug, user string) (*github.Membership, *github.Response, error)
	ListTeamReposBySlug(ctx context.Context, org, slug string, opts *github.ListOptions) ([]*github.Repository, *github.Response, error)
}

//counterfeiter:generate -o mocks/users_client.go -fake-name UsersClient . usersClient
type usersClient interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
}

//counterfeiter:generate -o mocks/repositories_client.go -fake-name RepositoriesClient . repositoriesClient
type repositoriesClient interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
}

type Config struct {
	Org            string `yaml:"org"`
	AppID          int64  `yaml:"appID"`
	InstallationID int64  `yaml:"installationID"`
	PrivateKey     string `yaml:"privateKey"`
	Logging        struct {
		Compression  bool   `yaml:"compression"`
		Ephemeral    bool   `yaml:"ephemeral"`
		LogDirectory string `yaml:"logDirectory"`
		MaxAge       int    `yaml:"maxAge"`
		MaxBackups   int    `yaml:"maxBackups"`
		MaxSize      int    `yaml:"maxSize"`
	} `yaml:"logging"`
}

type Manager struct {
	ActionsClient      actionsClient
	RepositoriesClient repositoriesClient
	TeamsClient        teamsClient

	Config *Config
	Logger *logrus.Logger

	CreateMaintainershipClient func(string) (*MaintainershipClient, error)
}

type MaintainershipClient struct {
	TeamsClient teamsClient
	UsersClient usersClient
}

type response struct {
	Message    string
	StatusCode int
}

// TODO: Add error paths and return errors
func (m *Manager) verifyMaintainership(token, team string) (bool, error) {
	client, err := m.CreateMaintainershipClient(token)
	if err != nil {
		return false, fmt.Errorf("failed retrieving user client: %w", err)
	}

	m.Logger.Info("Retrieving authorized user metadata")
	ctx := context.Background()
	user, _, err := client.UsersClient.Get(ctx, "")
	if err != nil {
		return false, fmt.Errorf("failed retrieving authenticated users data")
	}

	membership, resp, err := client.TeamsClient.GetTeamMembershipBySlug(ctx, m.Config.Org, team, user.GetLogin())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return false, fmt.Errorf("unable to locate team %s", team)
		}
		return false, err
	}
	return membership.GetRole() == "maintainer", nil
}

func (m *Manager) retrieveGroupID(name string) (*int64, error) {
	ctx := context.Background()
	groups, _, err := m.ActionsClient.ListOrganizationRunnerGroups(ctx, m.Config.Org, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed querying organization runner groups: %w", err)
	}
	for _, group := range groups.RunnerGroups {
		if group.GetName() == name {
			return group.ID, nil
		}
	}
	return nil, fmt.Errorf("unable to locate runner group with name %s", name)
}
