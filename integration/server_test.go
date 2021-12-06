/**
SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
	"github.com/google/uuid"
	"github.com/lindluni/actions-runner-manager/pkg/apis"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type Response struct {
	Code     int         `json:"code"`
	Error    string      `json:"error"`
	Response interface{} `json:"response"`
}

func initializeManager(t *testing.T) (*apis.Manager, []byte) {
	appID, err := strconv.Atoi(os.Getenv("MANAGER_APP_ID"))
	require.NoError(t, err)
	installationID, err := strconv.Atoi(os.Getenv("MANAGER_APP_INSTALLATION_ID"))
	require.NoError(t, err)
	encodedPrivateKey := os.Getenv("MANAGER_APP_PRIVATE_KEY")
	privateKey, err := base64.StdEncoding.DecodeString(encodedPrivateKey)
	require.NoError(t, err)

	config := &apis.Config{
		PrivateKey:     string(privateKey),
		AppID:          int64(appID),
		InstallationID: int64(installationID),
		Server: apis.Server{
			Address: "localhost",
			Port:    54321,
		},
		Org: os.Getenv("MANAGER_ORG"),
		Logging: apis.Logging{
			Ephemeral: true,
		},
	}
	logger, _ := test.NewNullLogger()
	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, config.InstallationID, privateKey)
	require.NoError(t, err)

	lmt := tollbooth.NewLimiter(5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetHeader("Authorization", []string{})
	lmt.SetHeaderEntryExpirationTTL(time.Hour)
	lmt.SetMessage(`{"code":429,"response":"You have reached maximum request limit. Please try again in a few seconds."}`)
	lmt.SetMessageContentType("application/json")

	createClient := func(token, uuid string) (*apis.MaintainershipClient, *github.User, error) {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		user, _, err := client.Users.Get(context.Background(), "")
		require.NoError(t, err)
		lmt.SetBasicAuthUsers(append(lmt.GetBasicAuthUsers(), user.GetLogin()))
		return &apis.MaintainershipClient{
			TeamsClient: client.Teams,
			UsersClient: client.Users,
		}, user, nil
	}

	router := gin.New()
	router.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	router.Use(gin.Logger())

	client := github.NewClient(&http.Client{Transport: itr})
	manager := &apis.Manager{
		ActionsClient:      client.Actions,
		RepositoriesClient: client.Repositories,
		TeamsClient:        client.Teams,
		Router:             router,
		Limit:              lmt,
		Server: &http.Server{
			Addr:    net.JoinHostPort(config.Server.Address, strconv.Itoa(config.Server.Port)),
			Handler: router,
		},
		Config:                     config,
		Logger:                     logger,
		CreateMaintainershipClient: createClient,
	}

	return manager, privateKey
}

func createGitHubClient(t *testing.T, config *apis.Config, privateKey []byte) *github.Client {
	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, config.InstallationID, privateKey)
	require.NoError(t, err)
	client := github.NewClient(&http.Client{Transport: itr})
	return client
}

func configureOrg(t *testing.T, slug string, client *github.Client, manager *apis.Manager) {
	repo, resp, err := client.Repositories.Create(context.Background(), manager.Config.Org, &github.Repository{
		Name: github.String(slug),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	team, resp, err := client.Teams.CreateTeam(context.Background(), manager.Config.Org, github.NewTeam{
		Name: slug,
		Maintainers: []string{
			os.Getenv("MANAGER_USER"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = client.Teams.AddTeamRepoBySlug(context.Background(), manager.Config.Org, team.GetSlug(), manager.Config.Org, repo.GetName(), &github.TeamAddTeamRepoOptions{})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// TODO: Replace with `/ping` endpoint
	time.Sleep(time.Second * 3)
}

func TestE2E(t *testing.T) {
	slug := fmt.Sprintf("integration_%s", uuid.NewString())
	manager, privateKey := initializeManager(t)
	manager.SetRoutes()
	client := createGitHubClient(t, manager.Config, privateKey)
	configureOrg(t, slug, client, manager)

	go func() {
		manager.Server.ListenAndServe()
	}()

	defer func() {
		resp, err := client.Teams.DeleteTeamBySlug(context.Background(), manager.Config.Org, slug)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, err = client.Repositories.Delete(context.Background(), manager.Config.Org, slug)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		err = manager.Server.Shutdown(context.Background())
		require.NoError(t, err)
	}()

	defer func() {
		groups, _, err := client.Actions.ListOrganizationRunnerGroups(context.Background(), manager.Config.Org, &github.ListOptions{PerPage: 100})
		require.NoError(t, err)
		var groupID int64
		for _, group := range groups.RunnerGroups {
			if group.GetName() == slug {
				groupID = group.GetID()
				break
			}
		}
		if groupID > 0 {
			_, err = client.Actions.DeleteOrganizationRunnerGroup(context.Background(), manager.Config.Org, groupID)
			require.NoError(t, err)
		}
	}()
	expected := &Response{
		Code:     http.StatusOK,
		Response: fmt.Sprintf("Runner group created successfully: %s", slug),
	}
	url := fmt.Sprintf("http://%s/api/v1/group-create?team=%s", manager.Server.Addr, slug)
	response := doGet(t, url)
	require.Equal(t, expected, response)

	expected = &Response{
		Code:     http.StatusOK,
		Response: "Successfully added repositories to runner group",
	}
	url = fmt.Sprintf("http://%s/api/v1/repos-add?team=%s&repos=%s", manager.Server.Addr, slug, slug)
	response = doGet(t, url)
	require.Equal(t, expected, response)

	expected = &Response{
		Code: http.StatusOK,
		Response: map[string]interface{}{
			"repos":   []interface{}{slug},
			"runners": []interface{}{},
		},
	}
	url = fmt.Sprintf("http://%s/api/v1/group-list?team=%s", manager.Server.Addr, slug)
	response = doGet(t, url)
	require.Equal(t, expected, response)

	expected = &Response{
		Code:     http.StatusOK,
		Response: "Successfully removed repositories from runner group",
	}
	url = fmt.Sprintf("http://%s/api/v1/repos-remove?team=%s&repos=%s", manager.Server.Addr, slug, slug)
	response = doGet(t, url)
	require.Equal(t, expected, response)

	expected = &Response{
		Code:     http.StatusOK,
		Response: "Successfully added repositories to runner group",
	}
	url = fmt.Sprintf("http://%s/api/v1/repos-set?team=%s&repos=%s", manager.Server.Addr, slug, slug)
	response = doGet(t, url)
	require.Equal(t, expected, response)

	expected = &Response{
		Code:     http.StatusOK,
		Response: fmt.Sprintf("Runner group deleted successfully: %s", slug),
	}
	url = fmt.Sprintf("http://%s/api/v1/group-delete?team=%s", manager.Server.Addr, slug)
	response = doGet(t, url)
	require.Equal(t, expected, response)

	url = fmt.Sprintf("http://%s/api/v1/token-register?team=%s", manager.Server.Addr, slug)
	response = doGet(t, url)
	require.Equal(t, http.StatusOK, response.Code)
	tokenMap := response.Response.(map[string]interface{})
	require.NotEmpty(t, tokenMap["token"])
	require.NotEmpty(t, tokenMap["expires_at"])

	url = fmt.Sprintf("http://%s/api/v1/token-remove?team=%s", manager.Server.Addr, slug)
	response = doGet(t, url)
	require.Equal(t, http.StatusOK, response.Code)
	tokenMap = response.Response.(map[string]interface{})
	require.NotEmpty(t, tokenMap["token"])
	require.NotEmpty(t, tokenMap["expires_at"])
}

func doGet(t *testing.T, url string) *Response {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	client := &http.Client{}
	req.Header.Set("Authorization", os.Getenv("MANAGER_ADMIN_PAT"))
	resp, err := client.Do(req)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var response *Response
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	return response
}
