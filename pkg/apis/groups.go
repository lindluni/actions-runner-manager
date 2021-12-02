package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	log "github.com/sirupsen/logrus"
)

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func (m *Manager) DoGroupCreate(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		log.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		log.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
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
			errMsg := "Runner group already exists: " + team
			http.Error(w, errMsg, http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("Runner group created successfully: %s", group.GetName()),
	})
	if err != nil {
		log.Error(err)
	}
}

func (m *Manager) DoGroupDelete(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		log.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		log.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	_, err = m.ActionsClient.DeleteOrganizationRunnerGroup(ctx, m.Config.Org, *id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("Runner group deleted successfully: %s", team),
	})
	if err != nil {
		log.Error(err)
	}
}

func (m *Manager) DoGroupList(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		log.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		log.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	runners, _, err := m.ActionsClient.ListRunnerGroupRunners(ctx, m.Config.Org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var filteredRunners []string
	for _, runner := range runners.Runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}

	repos, _, err := m.ActionsClient.ListRepositoryAccessRunnerGroup(ctx, m.Config.Org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(listResponse)
	if err != nil {
		log.Error(err)
	}
}
