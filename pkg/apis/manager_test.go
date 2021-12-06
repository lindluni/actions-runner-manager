/**
SPDX-License-Identifier: Apache-2.0
*/

package apis

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis/mocks"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

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

	logger, _ := test.NewNullLogger()
	for _, tc := range tests {
		client := &mocks.ActionsClient{}
		client.ListOrganizationRunnerGroupsReturnsOnCall(0, tc.input, &github.Response{NextPage: 1}, nil)
		client.ListOrganizationRunnerGroupsReturnsOnCall(1, tc.input, &github.Response{NextPage: 0}, nil)
		manager := &Manager{
			ActionsClient: client,
			Config:        &Config{},
			Logger:        logger,
		}
		id, _, err := manager.retrieveGroupID("fake-runner-group-name", "fake-uuid")
		require.NoError(t, err)
		require.Equal(t, tc.expected, id)
		require.Equal(t, client.ListOrganizationRunnerGroupsCallCount(), 2)
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

	logger, _ := test.NewNullLogger()
	for _, tc := range tests {
		client := &mocks.ActionsClient{}
		client.ListOrganizationRunnerGroupsReturns(tc.input, &github.Response{NextPage: 0, Response: &http.Response{StatusCode: http.StatusNotFound}}, tc.err)
		manager := &Manager{
			ActionsClient: client,
			Config:        &Config{},
			Logger:        logger,
		}
		id, statusCode, err := manager.retrieveGroupID("fake-runner-group-name", "fake-uuid")
		require.EqualError(t, err, tc.errString)
		require.Nil(t, tc.expected, id)
		require.Equal(t, statusCode, http.StatusNotFound)
		require.Equal(t, client.ListOrganizationRunnerGroupsCallCount(), 1)
	}
}

func TestVerifyMaintainership_Success(t *testing.T) {
	t.Parallel()

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
	isMaintainer, err := manager.verifyMaintainership("", "", "")
	require.NoError(t, err)
	require.False(t, isMaintainer)
}
