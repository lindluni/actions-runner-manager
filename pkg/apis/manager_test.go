package apis

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis/mocks"
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
		client.ListOrganizationRunnerGroupsReturns(tc.input, nil, nil)
		manager := &Manager{
			ActionsClient: client,
			Config:        &Config{},
			Logger:        logger,
		}
		id, err := manager.retrieveGroupID("fake-runner-group-name", "fake-uuid")
		require.NoError(t, err)
		require.Equal(t, tc.expected, id)
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

	logger, _ := test.NewNullLogger()
	for _, tc := range tests {
		client := &mocks.ActionsClient{}
		client.ListOrganizationRunnerGroupsReturns(tc.input, nil, tc.err)
		manager := &Manager{
			ActionsClient: client,
			Config:        &Config{},
			Logger:        logger,
		}
		id, err := manager.retrieveGroupID("fake-runner-group-name", "fake-uuid")
		require.EqualError(t, err, tc.errString)
		require.Nil(t, tc.expected, id)
		require.Equal(t, client.ListOrganizationRunnerGroupsCallCount(), 1)
	}
}
