// TODO: Do not panic in handlers
// TODO: Allow file logging or stdout logging or both via config
// TODO: Figure out a way to pull the org from the app or via config
// TODO: Implement better logging as a library?
// TODO: Implement pagination for github calls
// TODO: Add License and headers to all files
// TODO: Improve logging context
// TODO: Reimplement GETS as POSTS, this will require creating structs to marshal the body into
// TODO: Add CODEOWNERS and enforce it
// TODO: Write errors as response objects, not http calls
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const org = "department-of-veterans-affairs"

//go:generate counterfeiter -generate

//counterfeiter:generate -o mocks/actions_client.go -fake-name ActionsClient . actionsClient
type actionsClient interface {
	AddRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID, repoID int64) (*github.Response, error)
	CreateOrganizationRemoveToken(ctx context.Context, owner string) (*github.RemoveToken, *github.Response, error)
	CreateOrganizationRunnerGroup(ctx context.Context, org string, createReq github.CreateRunnerGroupRequest) (*github.RunnerGroup, *github.Response, error)
	CreateOrganizationRegistrationToken(ctx context.Context, owner string) (*github.RegistrationToken, *github.Response, error)
	DeleteOrganizationRunnerGroup(ctx context.Context, org string, groupID int64) (*github.Response, error)
	ListOrganizationRunnerGroups(ctx context.Context, org string, opts *github.ListOptions) (*github.RunnerGroups, *github.Response, error)
	ListRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID int64, opts *github.ListOptions) (*github.ListRepositories, *github.Response, error)
	ListRunnerGroupRunners(ctx context.Context, org string, groupID int64, opts *github.ListOptions) (*github.Runners, *github.Response, error)
	RemoveRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID, repoID int64) (*github.Response, error)
	SetRepositoryAccessRunnerGroup(ctx context.Context, org string, groupID int64, ids github.SetRepoAccessRunnerGroupRequest) (*github.Response, error)
}

type manager struct {
	actionsClient actionsClient
}

type response struct {
	StatusCode int
	Message    string
}

func createClient(token string) *github.Client {
	log.Info("Creating user GitHub client")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func verifyMaintainership(createClient func(token string) *github.Client, token, team string) bool {
	client := createClient(token)

	log.Info("Retrieving authorized user metadata")
	ctx := context.Background()
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Error("Failed retrieving user metadata")
	}

	membership, resp, err := client.Teams.GetTeamMembershipBySlug(ctx, org, team, user.GetLogin())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Fatalf("Unable to locate team %s", team)
		}
		log.Error(err)
	}
	return membership.GetRole() == "maintainer"
}

func (m *manager) retrieveGroupID(name string) (*int64, error) {
	ctx := context.Background()
	groups, _, err := m.actionsClient.ListOrganizationRunnerGroups(ctx, org, &github.ListOptions{PerPage: 100})
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

func (m *manager) doGroupAdd(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	group, resp, err := m.actionsClient.CreateOrganizationRunnerGroup(ctx, org, github.CreateRunnerGroupRequest{
		Name:                     github.String(team),
		Visibility:               github.String("private"),
		AllowsPublicRepositories: github.Bool(false),
	})
	if err != nil {
		if resp.StatusCode == http.StatusConflict {
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

func (m *manager) doGroupDelete(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	_, err = m.actionsClient.DeleteOrganizationRunnerGroup(ctx, org, *id)
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

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func (m *manager) doGroupList(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	runners, _, err := m.actionsClient.ListRunnerGroupRunners(ctx, org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var filteredRunners []string
	for _, runner := range runners.Runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}

	repos, _, err := m.actionsClient.ListRepositoryAccessRunnerGroup(ctx, org, *id, &github.ListOptions{PerPage: 100})
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

func (m *manager) doTokenRegister(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	token, _, err := m.actionsClient.CreateOrganizationRegistrationToken(ctx, org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		log.Error(err)
	}
}

func (m *manager) doTokenRemove(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	token, _, err := m.actionsClient.CreateOrganizationRemoveToken(ctx, org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		log.Error(err)
	}
}

func (m *manager) doReposAdd(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)
	tc := oauth2.NewClient(ctx, ts)
	userClient := github.NewClient(tc)

	repoIDs := map[string]int64{}
	for _, name := range repoNames {
		repo, resp, err := userClient.Repositories.Get(ctx, org, name)
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
		fmt.Println("Adding repo " + name + " to runner group " + team)
		_, err = m.actionsClient.AddRepositoryAccessRunnerGroup(ctx, org, *id, repoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&response{
		StatusCode: http.StatusOK,
		Message:    "Successfully added repositories to runner group",
	})
	if err != nil {
		log.Error(err)
	}
}

func (m *manager) doReposRemove(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)
	tc := oauth2.NewClient(ctx, ts)
	userClient := github.NewClient(tc)

	repoIDs := map[string]int64{}
	for _, name := range repoNames {
		repo, resp, err := userClient.Repositories.Get(ctx, org, name)
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
		log.Infof("Removing repo %s from runner group %s", name, team)
		_, err = m.actionsClient.RemoveRepositoryAccessRunnerGroup(ctx, org, *id, repoID)
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
		log.Error(err)
	}
}

func (m *manager) doReposSet(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(createClient, os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := m.retrieveGroupID(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)
	tc := oauth2.NewClient(ctx, ts)
	userClient := github.NewClient(tc)

	var repoIDs []int64
	for _, name := range repoNames {
		repo, resp, err := userClient.Repositories.Get(ctx, org, name)
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

	_, err = m.actionsClient.SetRepositoryAccessRunnerGroup(ctx, org, *id, github.SetRepoAccessRunnerGroupRequest{
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
		log.Error(err)
	}
}

func main() {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile("logs/server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))

	log.Info("Generating GitHub application credentials")
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 145597, 20164413, "key.pem")
	if err != nil {
		panic("Failed creating app authentication")
	}

	log.Info("Creating GitHub client")
	client := github.NewClient(&http.Client{Transport: itr})
	manager := &manager{
		actionsClient: client.Actions,
	}

	http.HandleFunc("/group-add", manager.doGroupAdd)
	http.HandleFunc("/group-delete", manager.doGroupDelete)
	http.HandleFunc("/group-list", manager.doGroupList)
	http.HandleFunc("/repos-add", manager.doReposAdd)
	http.HandleFunc("/repos-remove", manager.doReposRemove)
	http.HandleFunc("/repos-set", manager.doReposSet)
	http.HandleFunc("/token-register", manager.doTokenRegister)
	http.HandleFunc("/token-remove", manager.doTokenRemove)

	panic(http.ListenAndServe(":80", nil))
}
