/**
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/mocks"
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
