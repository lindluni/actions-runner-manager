/**
SPDX-License-Identifier: Apache-2.0
*/

package apis

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
)

// DoReposAdd    Add new repositories to an existing GitHub Actions organization runner group
// @Summary      Add new repositories to an existing GitHub Actions organization runner group
// @Description  Adds new repositories to an existing GitHub Actions organization named with the team slug
// @Tags         Repos
// @Produce      json
// @Param        team   query     string    true  "Canonical **slug** of the GitHub team"
// @Param        repos  query     []string  true  "Comma-seperated list of repository slugs"
// @Success      200    {object}  JSONResultSuccess{Code=int,Response=string}
// @Router       /repos-add [patch]
// @Security     ApiKeyAuth
func (m *Manager) DoReposAdd(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieving repo parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving repo parameter")
	repos := c.Query("repos")
	if repos == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: repos",
		})
		return
	}
	repoNames := strings.Split(repos, ",")
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieving repo parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, &JSONResultError{
			Code:  http.StatusUnauthorized,
			Error: "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group ID")
	groupID, statusCode, err := m.retrieveGroupID(team, uuid)
	if err != nil {
		c.JSON(statusCode, &JSONResultError{
			Code:  statusCode,
			Error: fmt.Sprintf("Unable to retrieve group ID: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Listing repositories assigned to team")
	repoIDs := map[string]int64{}
	var assignedRepos []*github.Repository
	opts := &github.ListOptions{PerPage: 100}
	for {
		teamRepos, resp, err := m.TeamsClient.ListTeamReposBySlug(ctx, m.Config.Org, team, opts)
		if err != nil {
			c.JSON(resp.StatusCode, &JSONResultError{
				Code:  resp.StatusCode,
				Error: fmt.Sprintf("Unable to retrieve team repos: %v", err),
			})
			return
		}
		assignedRepos = append(assignedRepos, teamRepos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Listed repositories assigned to team")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Mapping retrieved team repos to submitted repos")
	for _, name := range repoNames {
		m.Logger.Infof("Checking if team %s has access to repo %s", team, name)
		id, err := findRepoID(name, assignedRepos)
		if err != nil {
			c.JSON(http.StatusNotFound, &JSONResultError{
				Code:  http.StatusNotFound,
				Error: fmt.Sprintf("Repo %s not found in team %s: %v", name, team, err),
			})
			return
		}
		repoIDs[name] = id
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Mapped retrieved team repos to submitted repos")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Adding repositories to runner group")
	for name, repoID := range repoIDs {
		m.Logger.Infof("Adding repo %s to runner group %s", name, team)
		resp, err := m.ActionsClient.AddRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, repoID)
		if err != nil {
			c.JSON(resp.StatusCode, &JSONResultError{
				Code:  resp.StatusCode,
				Error: fmt.Sprintf("Unable to add repo %s to runner group %s: %v", name, team, err),
			})
			return
		}
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Added repositories to runner group")

	c.JSON(http.StatusOK, &JSONResultSuccess{
		Code:     http.StatusOK,
		Response: "Successfully added repositories to runner group",
	})
}

// DoReposRemove    Remove existing repositories from an existing GitHub Actions organization runner group
// @Summary      Remove existing repositories from an existing GitHub Actions organization runner group
// @Description  Removes existing repositories to an existing GitHub Actions organization named with the team slug
// @Tags         Repos
// @Produce      json
// @Param        team   query     string    true  "Canonical **slug** of the GitHub team"
// @Param        repos  query     []string  true  "Comma-seperated list of repository slugs"
// @Success      200    {object}  JSONResultSuccess{Code=int,Response=string}
// @Router       /repos-remove [patch]
// @Security     ApiKeyAuth
func (m *Manager) DoReposRemove(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving repos parameter")
	repos := c.Query("repos")
	if repos == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: repos",
		})
		return
	}
	repoNames := strings.Split(repos, ",")
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved repo parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, &JSONResultError{
			Code:  http.StatusUnauthorized,
			Error: "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group ID")
	groupID, statusCode, err := m.retrieveGroupID(team, uuid)
	if err != nil {
		c.JSON(statusCode, &JSONResultError{
			Code:  statusCode,
			Error: fmt.Sprintf("Unable to retrieve group ID: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving repository ID's")
	repoIDs := map[string]int64{}
	for _, name := range repoNames {
		repo, resp, err := m.RepositoriesClient.Get(ctx, m.Config.Org, name)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				c.JSON(resp.StatusCode, &JSONResultError{
					Code:  resp.StatusCode,
					Error: fmt.Sprintf("Repository %s not found", name),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, &JSONResultError{
				Code:  http.StatusInternalServerError,
				Error: fmt.Sprintf("Unable to retrieve repository %s: %v", name, err),
			})
			return
		}
		repoIDs[name] = repo.GetID()
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved repository ID's")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Removing repositories from runner group")
	for name, repoID := range repoIDs {
		m.Logger.Infof("Removing repo %s from runner group %s", name, team)
		resp, err := m.ActionsClient.RemoveRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, repoID)
		if err != nil {
			c.JSON(resp.StatusCode, &JSONResultError{
				Code:  resp.StatusCode,
				Error: fmt.Sprintf("Unable to remove repo %s from runner group %s: %v", name, team, err),
			})
			return
		}
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Removed repositories from runner group")

	c.JSON(http.StatusOK, &JSONResultSuccess{
		Code:     http.StatusOK,
		Response: "Successfully removed repositories from runner group",
	})
}

// DoReposSet       Replaces all existing repositories in an existing GitHub Actions organization runner group with a new set of repositories
// @Summary      Replaces all existing repositories in an existing GitHub Actions organization runner group with a new set of repositories
// @Description  Replaces all existing repositories in an existing GitHub Actions organization named with the team slug with a new set of repositories
// @Tags         Repos
// @Produce      json
// @Param        team   query     string    true  "Canonical **slug** of the GitHub team"
// @Param        repos  query     []string  true  "Comma-seperated list of repository slugs"
// @Success      200    {object}  JSONResultSuccess{Code=int,Response=string}
// @Router       /repos-set [patch]
// @Security     ApiKeyAuth
func (m *Manager) DoReposSet(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving assignedRepos parameter")
	repos := c.Query("repos")
	if repos == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: repos",
		})
		return
	}
	repoNames := strings.Split(repos, ",")
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved repo parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, &JSONResultError{
			Code:  http.StatusUnauthorized,
			Error: "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Listing repositories assigned to team")
	var assignedRepos []*github.Repository
	opts := &github.ListOptions{PerPage: 100}
	for {
		teamRepos, resp, err := m.TeamsClient.ListTeamReposBySlug(ctx, m.Config.Org, team, opts)
		if err != nil {
			c.JSON(resp.StatusCode, &JSONResultError{
				Code:  resp.StatusCode,
				Error: fmt.Sprintf("Unable to retrieve team assignedRepos: %v", err),
			})
			return
		}
		assignedRepos = append(assignedRepos, teamRepos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Listed repositories assigned to team")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Mapping retrieved team assignedRepos to submitted assignedRepos")
	var repoIDs []int64
	for _, name := range repoNames {
		m.Logger.Infof("Checking if team %s has access to repo %s", team, name)
		id, err := findRepoID(name, assignedRepos)
		if err != nil {
			c.JSON(http.StatusNotFound, &JSONResultError{
				Code:  http.StatusNotFound,
				Error: fmt.Sprintf("Repo %s not found in team %s: %v", name, team, err),
			})
			return
		}
		repoIDs = append(repoIDs, id)
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Mapped retrieved team assignedRepos to submitted assignedRepos")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving runner group ID")
	groupID, statusCode, err := m.retrieveGroupID(team, uuid)
	if err != nil {
		c.JSON(statusCode, &JSONResultError{
			Code:  statusCode,
			Error: fmt.Sprintf("Unable to retrieve group ID: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved runner group ID")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Adding repositories to runner group")
	resp, err := m.ActionsClient.SetRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, github.SetRepoAccessRunnerGroupRequest{
		SelectedRepositoryIDs: repoIDs,
	})
	if err != nil {
		c.JSON(resp.StatusCode, &JSONResultError{
			Code:  resp.StatusCode,
			Error: fmt.Sprintf("Unable to set repositories for runner group %s: %v", team, err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Added repositories to runner group")

	c.JSON(http.StatusOK, &JSONResultSuccess{
		Code:     http.StatusOK,
		Response: "Successfully added repositories to runner group",
	})
}

func findRepoID(name string, teamRepos []*github.Repository) (int64, error) {
	for _, teamRepo := range teamRepos {
		if name == teamRepo.GetName() {
			return teamRepo.GetID(), nil
		}
	}
	return -1, fmt.Errorf("team does not have repo access")
}
