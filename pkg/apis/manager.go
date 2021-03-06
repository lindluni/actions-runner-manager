/**
SPDX-License-Identifier: Apache-2.0
*/

package apis

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
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
	Org            string  `yaml:"org"`
	AppID          int64   `yaml:"appID"`
	InstallationID int64   `yaml:"installationID"`
	PrivateKey     string  `yaml:"privateKey"`
	Logging        Logging `yaml:"logging"`
	Server         Server  `yaml:"server"`
}

type Logging struct {
	Compression  bool   `yaml:"compression"`
	Ephemeral    bool   `yaml:"ephemeral"`
	Level        string `yaml:"level"`
	LogDirectory string `yaml:"logDirectory"`
	MaxAge       int    `yaml:"maxAge"`
	MaxBackups   int    `yaml:"maxBackups"`
	MaxSize      int    `yaml:"maxSize"`
}

type Server struct {
	Address   string  `yaml:"address"`
	Port      int     `yaml:"port"`
	RateLimit float64 `yaml:"rateLimit"`
	TLS       TLS     `yaml:"tls"`
}

type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type Manager struct {
	ActionsClient      actionsClient
	RepositoriesClient repositoriesClient
	TeamsClient        teamsClient

	Limit  *limiter.Limiter
	Router *gin.Engine
	Server *http.Server

	Config *Config
	Logger *logrus.Logger

	CreateMaintainershipClient func(string, string) (*MaintainershipClient, *github.User, error)
}

type MaintainershipClient struct {
	TeamsClient teamsClient
	UsersClient usersClient
}

func (m *Manager) Serve() {
	m.Logger.Info("Initializing API endpoints")
	m.SetRoutes()

	m.Logger.Info("Configuring OS signal handling")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		err := m.Server.Shutdown(context.Background())
		m.Logger.Errorf("Failed to shutdown server: %v", err)
	}()
	m.Logger.Debug("Configured OS signal handling")

	m.Logger.Debug("Compiling HTTP server address")
	address := fmt.Sprintf("%s:%d", m.Config.Server.Address, m.Config.Server.Port)
	m.Logger.Infof("Starting API server on address: %s", address)
	if m.Config.Server.TLS.Enabled {
		err := m.Server.ListenAndServeTLS(m.Config.Server.TLS.CertFile, m.Config.Server.TLS.KeyFile)
		if err != nil {
			m.Logger.Fatalf("API server failed: %v", err)
		}
	} else {
		err := m.Server.ListenAndServe()
		if err != nil {
			m.Logger.Fatalf("API server failed: %v", err)
		}
	}
}

func (m *Manager) SetRoutes() {
	v1 := m.Router.Group("/api/v1")
	{
		v1.POST("/group-create", LimitHandler(m.Limit), m.DoGroupCreate)
		v1.DELETE("/group-delete", LimitHandler(m.Limit), m.DoGroupDelete)
		v1.GET("/group-list", LimitHandler(m.Limit), m.DoGroupList)
		v1.PATCH("/repos-add", LimitHandler(m.Limit), m.DoReposAdd)
		v1.PATCH("/repos-remove", LimitHandler(m.Limit), m.DoReposRemove)
		v1.PATCH("/repos-set", LimitHandler(m.Limit), m.DoReposSet)
		v1.GET("/token-register", LimitHandler(m.Limit), m.DoTokenRegister)
		v1.GET("/token-remove", LimitHandler(m.Limit), m.DoTokenRemove)
		v1.GET("/status", LimitHandler(m.Limit), m.Status)
	}
	m.Logger.Debug("Initialized API endpoints")
}

func (m *Manager) Status(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.WithField("uuid", uuid).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).Debug("Retrieved Authorization header")

	_, _, _ = m.CreateMaintainershipClient(token, uuid)
	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": "Server ready",
	})
}

func LimitHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func (m *Manager) verifyMaintainership(token, team, uuid string) (bool, error) {
	m.Logger.WithField("uuid", uuid).Info("Creating maintainership client")
	client, user, err := m.CreateMaintainershipClient(token, uuid)
	if err != nil {
		return false, fmt.Errorf("failed retrieving user client: %w", err)
	}
	m.Logger.WithField("uuid", uuid).Debug("Created maintainership client")

	m.Logger.WithField("uuid", uuid).Infof("Retrieving team: %s", team)
	membership, resp, err := client.TeamsClient.GetTeamMembershipBySlug(context.Background(), m.Config.Org, team, user.GetLogin())
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, fmt.Errorf("unable to locate team %s", team)
		}
		return false, err
	}
	m.Logger.WithField("uuid", uuid).Debugf("Retrieved team %s", team)

	return membership.GetRole() == "maintainer", nil
}

func (m *Manager) retrieveGroupID(name, uuid string) (*int64, int, error) {
	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).Info("Retrieving runner groups")
	var groups []*github.RunnerGroup
	opts := &github.ListOptions{PerPage: 100}
	for {
		runnerGroups, resp, err := m.ActionsClient.ListOrganizationRunnerGroups(ctx, m.Config.Org, opts)
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("failed querying organization runner groups: %w", err)
		}
		groups = append(groups, runnerGroups.RunnerGroups...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", uuid).Debug("Retrieved runner groups")

	m.Logger.WithField("uuid", uuid).Info("Searching for runner group")
	for _, group := range groups {
		if group.GetName() == name {
			m.Logger.WithField("uuid", uuid).Debug("Found runner group")
			return group.ID, http.StatusOK, nil
		}
	}

	return nil, http.StatusNotFound, fmt.Errorf("unable to locate runner group with name %s", name)
}
