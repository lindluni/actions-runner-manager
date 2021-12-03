package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	"github.com/google/uuid"
)

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func (m *Manager) DoGroupCreate(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, id)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Creating runner group")
	group, resp, err := m.ActionsClient.CreateOrganizationRunnerGroup(ctx, m.Config.Org, github.CreateRunnerGroupRequest{
		Name:                     github.String(team),
		Visibility:               github.String("selected"),
		AllowsPublicRepositories: github.Bool(false),
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusConflict {
			message := fmt.Sprintf("Runner group already exists: %s", team)
			m.writeResponse(w, http.StatusConflict, message)
			return
		}
		message := fmt.Sprintf("Unable to create runner group: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Created runner group")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   fmt.Sprintf("Runner group created successfully: %s", group.GetName()),
	}
	m.writeResponseWithUUID(w, response, id)
}

func (m *Manager) DoGroupDelete(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, id)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving runner group ID")
	groupID, err := m.retrieveGroupID(team, id)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve group ID: %v", err))
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Deleting runner group")
	_, err = m.ActionsClient.DeleteOrganizationRunnerGroup(ctx, m.Config.Org, *groupID)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to delete runner group: %v", err))
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Deleted runner group")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   fmt.Sprintf("Runner group deleted successfully: %s", team),
	}
	m.writeResponseWithUUID(w, response, id)
}

func (m *Manager) DoGroupList(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, id)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Verified maintainership")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving runner group ID")
	groupID, err := m.retrieveGroupID(team, id)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve group ID: %v", err))
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving runner group runner list")
	var runners []*github.Runner
	opts := &github.ListOptions{PerPage: 100}
	for {
		runnerGroupRunners, resp, err := m.ActionsClient.ListRunnerGroupRunners(ctx, m.Config.Org, *groupID, opts)
		if err != nil {
			m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list runners: %v", err))
			return
		}
		runners = append(runners, runnerGroupRunners.Runners...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group runner list")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Generating runner list")
	var filteredRunners []string
	for _, runner := range runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Generated runner list")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving runner group repository list")
	var repos []*github.Repository
	opts = &github.ListOptions{PerPage: 100}
	for {
		runnerGroupRepos, resp, err := m.ActionsClient.ListRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, opts)
		if err != nil {
			m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list repositories: %v", err))
			return
		}
		repos = append(repos, runnerGroupRepos.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group repository list")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Generating repository list")
	var filteredRepos []string
	for _, repo := range repos {
		filteredRepos = append(filteredRepos, repo.GetName())
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Generated repository list")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Generating response")
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
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Generated response")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   listResponse,
	}
	m.writeResponseWithUUID(w, response, id)
}
