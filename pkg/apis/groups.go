package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
)

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func (m *Manager) DoGroupCreate(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}

	ctx := context.Background()
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

	m.writeResponse(w, http.StatusOK, fmt.Sprintf("Runner group created successfully: %s", group.GetName()))
}

func (m *Manager) DoGroupDelete(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve group ID: %v", err))
		return
	}

	ctx := context.Background()
	_, err = m.ActionsClient.DeleteOrganizationRunnerGroup(ctx, m.Config.Org, *id)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to delete runner group: %v", err))
		return
	}

	m.writeResponse(w, http.StatusOK, fmt.Sprintf("Runner group deleted successfully: %s", team))
}

func (m *Manager) DoGroupList(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusForbidden, "Missing Authorization header")
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to retrieve group ID: %v", err))
		return
	}

	ctx := context.Background()
	runners, _, err := m.ActionsClient.ListRunnerGroupRunners(ctx, m.Config.Org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list runners: %v", err))
		return
	}
	var filteredRunners []string
	for _, runner := range runners.Runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}

	repos, _, err := m.ActionsClient.ListRepositoryAccessRunnerGroup(ctx, m.Config.Org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to list repositories: %v", err))
		return
	}
	var filteredRepos []string
	for _, repo := range repos.Repositories {
		filteredRepos = append(filteredRepos, repo.GetName())
	}

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

	message, err := json.Marshal(listResponse)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to marshal response: %v", err))
		return
	}

	m.writeResponse(w, http.StatusOK, string(message))
}
