package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const org = "department-of-veterans-affairs"

var appClient *github.Client

func init() {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 145597, 20164413, "key.pem")
	if err != nil {
		panic("Failed creating app authentication")
	}

	appClient = github.NewClient(&http.Client{Transport: itr})
}

func verifyMaintainership(token, team string) bool {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}

	membership, resp, err := client.Teams.GetTeamMembershipBySlug(ctx, org, team, user.GetLogin())
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			panic(fmt.Errorf("team not found"))
		}
		panic(err)
	}
	if membership.GetRole() == "maintainer" {
		return true
	}
	return false
}

func retrieveGroupID(client *github.Client, name string) (*int64, error) {
	ctx := context.Background()
	groups, _, err := client.Actions.ListOrganizationRunnerGroups(ctx, org, &github.ListOptions{PerPage: 100})
	if err != nil {
		panic(err)
	}
	for _, group := range groups.RunnerGroups {
		if group.GetName() == name {
			return group.ID, nil
		}
	}
	return nil, fmt.Errorf("unable to locate runner group")
}

func doGroupAdd(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	group, resp, err := appClient.Actions.CreateOrganizationRunnerGroup(ctx, org, github.CreateRunnerGroupRequest{
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
	_, _ = fmt.Fprintf(w, "Runner group created successfully: %s", group.GetName())
}

func doGroupDelete(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := retrieveGroupID(appClient, team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	_, err = appClient.Actions.DeleteOrganizationRunnerGroup(ctx, org, *id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "Runner group deleted successfully: %s", team)
}

type listResponse struct {
	Repos   []string `json:"repos"`
	Runners []string `json:"runners"`
}

func doGroupList(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	id, err := retrieveGroupID(appClient, team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	runners, _, err := appClient.Actions.ListRunnerGroupRunners(ctx, org, *id, &github.ListOptions{PerPage: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var filteredRunners []string
	for _, runner := range runners.Runners {
		filteredRunners = append(filteredRunners, runner.GetName())
	}

	repos, _, err := appClient.Actions.ListRepositoryAccessRunnerGroup(ctx, org, *id, &github.ListOptions{PerPage: 100})
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

	response, err := json.Marshal(listResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(response))
	_, _ = fmt.Fprintf(w, string(response))
}

func doTokenRegister(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	token, _, err := appClient.Actions.CreateOrganizationRegistrationToken(ctx, org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, string(response))
}

func doTokenRemove(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	ctx := context.Background()
	token, _, err := appClient.Actions.CreateOrganizationRemoveToken(ctx, org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, string(response))
}

func doReposAdd(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 145597, 20164413, "key.pem")
	if err != nil {
		panic("Failed creating app authentication")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	id, err := retrieveGroupID(client, team)
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
		_, err = client.Actions.AddRepositoryAccessRunnerGroup(ctx, org, *id, repoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	fmt.Fprintf(w, "Successfully added repositories to runner group")
}

func doReposRemove(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 145597, 20164413, "key.pem")
	if err != nil {
		panic("Failed creating app authentication")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	id, err := retrieveGroupID(client, team)
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
		fmt.Println("Removing repo " + name + " to runner group " + team)
		_, err = client.Actions.RemoveRepositoryAccessRunnerGroup(ctx, org, *id, repoID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	fmt.Fprintf(w, "Successfully removed repositories to runner group")
}

func doReposSet(w http.ResponseWriter, req *http.Request) {
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

	isMaintainer := verifyMaintainership(os.Getenv("GITHUB_PAT"), team)
	if !isMaintainer {
		panic("User is not a maintainer of the team")
	}

	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 145597, 20164413, "key.pem")
	if err != nil {
		panic("Failed creating app authentication")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	id, err := retrieveGroupID(client, team)
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

	_, err = client.Actions.SetRepositoryAccessRunnerGroup(ctx, org, *id, github.SetRepoAccessRunnerGroupRequest{
		SelectedRepositoryIDs: repoIDs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Successfully replaced all repositories in runner group")
}

func main() {
	http.HandleFunc("/group-add", doGroupAdd)
	http.HandleFunc("/group-delete", doGroupDelete)
	http.HandleFunc("/group-list", doGroupList)
	http.HandleFunc("/repos-add", doReposAdd)
	http.HandleFunc("/repos-remove", doReposRemove)
	http.HandleFunc("/repos-set", doReposSet)
	http.HandleFunc("/token-register", doTokenRegister)
	http.HandleFunc("/token-remove", doTokenRemove)

	panic(http.ListenAndServe(":80", nil))
}
