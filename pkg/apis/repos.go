package apis

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/google/uuid"
)

func (m *Manager) DoReposAdd(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieving repo parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving repo parameter")
	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: repos")
		return
	}
	repoNames := strings.Split(reposParam[0], ",")
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieving repo parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing Authorization header")
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
		message := fmt.Sprintf("Unable to retrieve group ID: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Listing repositories assigned to team")
	repoIDs := map[string]int64{}
	teamRepos, _, err := m.TeamsClient.ListTeamReposBySlug(ctx, m.Config.Org, team, &github.ListOptions{PerPage: 100})
	if err != nil {
		message := fmt.Sprintf("Unable to retrieve team repos: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Listed repositories assigned to team")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Mapping retrieved team repos to submitted repos")
	for _, name := range repoNames {
		m.Logger.Infof("Checking if team %s has access to repo %s", team, name)
		id, err := findRepoID(name, teamRepos)
		if err != nil {
			message := fmt.Sprintf("Repo %s not found in team %s: %v", name, team, err)
			m.writeResponse(w, http.StatusNotFound, message)
			return
		}
		repoIDs[name] = id
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Mapped retrieved team repos to submitted repos")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Adding repositories to runner group")
	for name, repoID := range repoIDs {
		m.Logger.Infof("Adding repo %s to runner group %s", name, team)
		_, err = m.ActionsClient.AddRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, repoID)
		if err != nil {
			message := fmt.Sprintf("Unable to add repo %s to runner group %s: %v", name, team, err)
			m.writeResponse(w, http.StatusInternalServerError, message)
			return
		}
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Added repositories to runner group")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   "Successfully added repositories to runner group",
	}
	m.writeResponseWithUUID(w, response, id)
}

func (m *Manager) DoReposRemove(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving repos parameter")
	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: repos")
		return
	}
	repoNames := strings.Split(reposParam[0], ",")
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved repo parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing Authorization header")
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
		message := fmt.Sprintf("Unable to retrieve group ID: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving repository ID's")
	repoIDs := map[string]int64{}
	for _, name := range repoNames {
		repo, resp, err := m.RepositoriesClient.Get(ctx, m.Config.Org, name)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				message := fmt.Sprintf("Repository %s not found", name)
				m.writeResponse(w, http.StatusNotFound, message)
				return
			}
			message := fmt.Sprintf("Unable to retrieve repository %s: %v", name, err)
			m.writeResponse(w, http.StatusInternalServerError, message)
			return
		}
		repoIDs[name] = repo.GetID()
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved repository ID's")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Removing repositories from runner group")
	for name, repoID := range repoIDs {
		m.Logger.Infof("Removing repo %s from runner group %s", name, team)
		_, err = m.ActionsClient.RemoveRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, repoID)
		if err != nil {
			message := fmt.Sprintf("Unable to remove repo %s from runner group %s: %v", name, team, err)
			m.writeResponse(w, http.StatusInternalServerError, message)
		}
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Removed repositories from runner group")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   "Successfully removed repositories from runner group",
	}
	m.writeResponseWithUUID(w, response, id)
}

func (m *Manager) DoReposSet(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving repos parameter")
	reposParam := req.URL.Query()["repos"]
	if len(reposParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: repos")
		return
	}
	repoNames := strings.Split(reposParam[0], ",")
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved repo parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing Authorization header")
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
		message := fmt.Sprintf("Unable to retrieve group ID: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved runner group ID")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving repository ID's")
	var repoIDs []int64
	for _, name := range repoNames {
		repo, resp, err := m.RepositoriesClient.Get(ctx, m.Config.Org, name)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				message := fmt.Sprintf("Repository %s not found", name)
				m.writeResponse(w, http.StatusNotFound, message)
				return
			}
			message := fmt.Sprintf("Unable to retrieve repository %s: %v", name, err)
			m.writeResponse(w, http.StatusInternalServerError, message)
			return
		}
		repoIDs = append(repoIDs, repo.GetID())
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved repository ID's")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Adding repositories to runner group")
	_, err = m.ActionsClient.SetRepositoryAccessRunnerGroup(ctx, m.Config.Org, *groupID, github.SetRepoAccessRunnerGroupRequest{
		SelectedRepositoryIDs: repoIDs,
	})
	if err != nil {
		message := fmt.Sprintf("Unable to set repositories for runner group %s: %v", team, err)
		m.writeResponse(w, http.StatusInternalServerError, message)
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Added repositories to runner group")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   "Successfully added repositories to runner group",
	}
	m.writeResponseWithUUID(w, response, id)
}

func findRepoID(name string, teamRepos []*github.Repository) (int64, error) {
	for _, teamRepo := range teamRepos {
		if name == teamRepo.GetName() {
			return teamRepo.GetID(), nil
		}
	}
	return -1, fmt.Errorf("team does not have repo access")
}
