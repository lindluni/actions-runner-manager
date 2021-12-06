/**
SPDX-License-Identifier: Apache-2.0
*/

package apis

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis/mocks"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestDoGroupCreate_Success(t *testing.T) {
	actionsClient := &mocks.ActionsClient{}
	teamsClient := &mocks.TeamsClient{}
	usersClient := &mocks.UsersClient{}
	logger, _ := test.NewNullLogger()
	manager := &Manager{
		ActionsClient: actionsClient,
		Config:        &Config{},
		CreateMaintainershipClient: func(string, string) (*MaintainershipClient, *github.User, error) {
			return &MaintainershipClient{
				TeamsClient: teamsClient,
				UsersClient: usersClient,
			}, nil, nil
		},
		Logger: logger,
	}

	runnerGroup := &github.RunnerGroup{
		Name: github.String("fake-runner-group-name"),
	}
	actionsClient.CreateOrganizationRunnerGroupReturns(runnerGroup, nil, nil)

	membership := &github.Membership{
		Role: github.String("maintainer"),
	}
	teamsClient.GetTeamMembershipBySlugReturns(membership, nil, nil)

	var err error
	writer := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(writer)
	context.Request, err = http.NewRequest(http.MethodGet, "/v1/api/group-add?team=fake-team", nil)
	context.Request.Header.Set("Authorization", "test-token")
	require.NoError(t, err)

	manager.DoGroupCreate(context)
	result := writer.Result()
	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)

	expected := &gin.H{
		"Code":     float64(http.StatusOK),
		"Response": "Runner group created successfully: fake-runner-group-name",
	}

	groupAddResponse := &gin.H{}
	err = json.Unmarshal(body, &groupAddResponse)
	require.NoError(t, err)
	require.Equal(t, expected, groupAddResponse)
	require.Equal(t, 1, actionsClient.CreateOrganizationRunnerGroupCallCount())
	require.Equal(t, 1, teamsClient.GetTeamMembershipBySlugCallCount())
}
