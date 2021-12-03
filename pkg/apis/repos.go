package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v41/github"
)

func (m *Manager) DoReposAdd(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		http.Error(w, "Missing required parameter: repos", http.StatusBadRequest)
		return
	}
	repoNames := strings.Split(reposParam[0], ",")

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		m.Logger.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		m.Logger.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	repoIDs := map[string]int64{}
	teamRepos, _, err := m.TeamsClient.ListTeamReposBySlug(ctx, m.Config.Org, team, &github.ListOptions{PerPage: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, name := range repoNames {
		m.Logger.Infof("Checking if team %s has access to repo %s", team, name)
		id, err := findRepoID(name, teamRepos)
		if err != nil {
			m.Logger.Errorf("Team %s has no access to repo %s", team, name)
			http.Error(w, fmt.Sprintf("Repo %s not found in team %s: %v", name, team, err), http.StatusNotFound)
			return
		}
		repoIDs[name] = id
	}

	for name, repoID := range repoIDs {
		m.Logger.Infof("Adding repo %s to runner group %s", name, team)
		_, err = m.ActionsClient.AddRepositoryAccessRunnerGroup(ctx, m.Config.Org, *id, repoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    "Successfully added repositories to runner group",
	})
	if err != nil {
		m.Logger.Error(err)
	}
}

func (m *Manager) DoReposRemove(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		http.Error(w, "Missing required parameter: repos", http.StatusBadRequest)
		return
	}
	repoNames := strings.Split(reposParam[0], ",")

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		m.Logger.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		m.Logger.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	repoIDs := map[string]int64{}
	for _, name := range repoNames {
		repo, resp, err := m.RepositoriesClient.Get(ctx, m.Config.Org, name)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				http.Error(w, "Repository not found: "+name, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		repoIDs[name] = repo.GetID()
	}

	for name, repoID := range repoIDs {
		m.Logger.Infof("Removing repo %s from runner group %s", name, team)
		_, err = m.ActionsClient.RemoveRepositoryAccessRunnerGroup(ctx, m.Config.Org, *id, repoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    "Successfully removed repositories from runner group",
	})
	if err != nil {
		m.Logger.Error(err)
	}
}

func (m *Manager) DoReposSet(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		http.Error(w, "Missing required parameter: repos", http.StatusBadRequest)
		return
	}
	repoNames := strings.Split(reposParam[0], ",")

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		m.Logger.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		m.Logger.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	var repoIDs []int64
	for _, name := range repoNames {
		repo, resp, err := m.RepositoriesClient.Get(ctx, m.Config.Org, name)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				http.Error(w, "Repository not found: "+name, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		repoIDs = append(repoIDs, repo.GetID())
	}

	_, err = m.ActionsClient.SetRepositoryAccessRunnerGroup(ctx, m.Config.Org, *id, github.SetRepoAccessRunnerGroupRequest{
		SelectedRepositoryIDs: repoIDs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    "Successfully replaced all repositories in runner group",
	})
	if err != nil {
		m.Logger.Error(err)
	}
}

func findRepoID(name string, teamRepos []*github.Repository) (int64, error) {
	for _, teamRepo := range teamRepos {
		if name == teamRepo.GetName() {
			return teamRepo.GetID(), nil
		}
	}
	return -1, fmt.Errorf("team does not have repo access")
}
