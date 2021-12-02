/**
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/mocks"
	"github.com/stretchr/testify/require"
)

func TestDoGroupAdd_Success(t *testing.T) {
	actionsClient := &mocks.ActionsClient{}
	teamsClient := &mocks.TeamsClient{}
	usersClient := &mocks.UsersClient{}
	manager := &manager{
		actionsClient: actionsClient,
		createMaintainershipClient: func(s string) *maintainershipClient {
			return &maintainershipClient{
				teamsClient: teamsClient,
				usersClient: usersClient,
			}
		},
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

	manager.doGroupAdd(writer, request)
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

func TestRetrieveGroupID_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    *github.RunnerGroups
		expected *int64
	}{
		{
			input: &github.RunnerGroups{
				RunnerGroups: []*github.RunnerGroup{
					{
						Name: github.String("fake-runner-group-name"),
						ID:   github.Int64(1000),
					},
				},
			},
			expected: github.Int64(1000),
		},
	}

	for _, test := range tests {
		client := &mocks.ActionsClient{}
		client.ListOrganizationRunnerGroupsReturns(test.input, nil, nil)
		manager := &manager{actionsClient: client}
		id, err := manager.retrieveGroupID("fake-runner-group-name")
		require.NoError(t, err)
		require.Equal(t, test.expected, id)
		require.Equal(t, client.ListOrganizationRunnerGroupsCallCount(), 1)
	}
}

func TestRetrieveGroupID_Failure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err       error
		errString string
		expected  *int64
		input     *github.RunnerGroups
	}{
		{
			input: &github.RunnerGroups{
				RunnerGroups: []*github.RunnerGroup{
					{
						Name: github.String("unknown-runner-group-name"),
						ID:   github.Int64(1000),
					},
				},
			},
			err:       nil,
			errString: "unable to locate runner group with name fake-runner-group-name",
			expected:  nil,
		},
		{
			input: &github.RunnerGroups{
				RunnerGroups: []*github.RunnerGroup{
					{
						Name: github.String("unknown-runner-group-name"),
						ID:   github.Int64(1000),
					},
				},
			},
			err:       fmt.Errorf("fake-error"),
			errString: "failed querying organization runner groups: fake-error",
			expected:  nil,
		},
	}

	for _, test := range tests {
		client := &mocks.ActionsClient{}
		client.ListOrganizationRunnerGroupsReturns(test.input, nil, test.err)
		manager := &manager{actionsClient: client}
		id, err := manager.retrieveGroupID("fake-runner-group-name")
		require.EqualError(t, err, test.errString)
		require.Nil(t, test.expected, id)
		require.Equal(t, client.ListOrganizationRunnerGroupsCallCount(), 1)
	}
}
