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

	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis/mocks"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestDoGroupAdd_Success(t *testing.T) {
	actionsClient := &mocks.ActionsClient{}
	teamsClient := &mocks.TeamsClient{}
	usersClient := &mocks.UsersClient{}
	logger, _ := test.NewNullLogger()
	manager := &Manager{
		ActionsClient: actionsClient,
		Config:        &Config{},
		CreateMaintainershipClient: func(s string) (*MaintainershipClient, error) {
			return &MaintainershipClient{
				TeamsClient: teamsClient,
				UsersClient: usersClient,
			}, nil
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

	writer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/group-add?team=fake-team", nil)
	request.Header.Set("AUTHORIZATION", "test-token")

	manager.DoGroupCreate(writer, request)
	result := writer.Result()
	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)

	expected := &response{
		StatusCode: http.StatusOK,
		Message:    "Runner group created successfully: fake-runner-group-name",
	}

	groupAddResponse := &response{}
	err = json.Unmarshal(body, &groupAddResponse)
	require.NoError(t, err)
	require.Equal(t, expected, groupAddResponse)
	require.Equal(t, actionsClient.CreateOrganizationRunnerGroupCallCount(), 1)
	require.Equal(t, teamsClient.GetTeamMembershipBySlugCallCount(), 1)
	require.Equal(t, usersClient.GetCallCount(), 1)
}
