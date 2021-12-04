package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
)

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func (m *Manager) DoGroupCreate(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info(c, "Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"error": "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Creating runner group")
	group, resp, err := m.ActionsClient.CreateOrganizationRunnerGroup(ctx, m.Config.Org, github.CreateRunnerGroupRequest{
		Name:                     github.String(team),
		Visibility:               github.String("selected"),
		AllowsPublicRepositories: github.Bool(false),
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{
				"Code":  http.StatusConflict,
				"error": fmt.Sprintf("Runner group already exists: %s", team),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"Code":  http.StatusInternalServerError,
			"error": fmt.Sprintf("Unable to create runner group: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created runner group")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": fmt.Sprintf("Runner group created successfully: %s", group.GetName()),
	})
}

func (m *Manager) DoGroupDelete(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info(c, "Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"error": "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group ID")
	groupID, statusCode, err := m.retrieveGroupID(team, uuid)
	if err != nil {
		c.JSON(statusCode, gin.H{
			"Code":  statusCode,
			"error": fmt.Sprintf("Unable to retrieve group ID: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Deleting runner group")
	resp, err := m.ActionsClient.DeleteOrganizationRunnerGroup(ctx, m.Config.Org, *groupID)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{
			"Code":  resp.StatusCode,
			"error": fmt.Sprintf("Unable to delete runner group: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Deleted runner group")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": fmt.Sprintf("Runner group deleted successfully: %s", team),
	})
}

func (m *Manager) DoGroupList(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info(c, "Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"error": "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group ID")
	groupID, statusCode, err := m.retrieveGroupID(team, uuid)
	if err != nil {
		c.JSON(statusCode, gin.H{
			"Code":  statusCode,
			"error": fmt.Sprintf("Unable to retrieve group ID: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group runner list")
	var runners []*github.Runner
	opts := &github.ListOptions{PerPage: 100}
	for {
		runnerGroupRunners, resp, err := m.ActionsClient.ListRunnerGroupRunners(ctx, m.Config.Org, *groupID, opts)
		if err != nil {
			c.JSON(resp.StatusCode, gin.H{
				"Code":  resp.StatusCode,
				"error": fmt.Sprintf("Unable to list runners: %v", err),
			})
			return
		}
		runners = append(runners, runnerGroupRunners.Runners...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group runner list")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Generating runner list")
	var filteredRunners []string
	for _, runner := range runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Generated runner list")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group repository list")
	var repos []*github.Repository
	opts = &github.ListOptions{PerPage: 100}
	for {
		runnerGroupRepos, resp, err := m.ActionsClient.ListRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Code":  http.StatusInternalServerError,
				"error": fmt.Sprintf("Unable to list repositories: %v", err),
			})
			return
		}
		repos = append(repos, runnerGroupRepos.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group repository list")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Generating repository list")
	var filteredRepos []string
	for _, repo := range repos {
		filteredRepos = append(filteredRepos, repo.GetName())
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Generated repository list")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Generating Response")
	listResponse := &listResponse{
		Repos:   filteredRepos,
		Runners: filteredRunners,
	}
	if listResponse.Repos == nil {
		listResponse.Repos = []string{}
	}
	if listResponse.Runners == nil {
		listResponse.Runners = []string{}
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug(c, "Generated Response")

	c.JSON(http.StatusForbidden, gin.H{
		"Code":     http.StatusForbidden,
		"Response": listResponse,
	})
}
