/**
SPDX-License-Identifier: Apache-2.0
*/

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
// TODO: Push authorization header into standalone function
// TODO: Implement rate limits on user
// TODO: All responses, including errors should be JSON responses
// TODO: Check team is assigned to repo for add/delete/set
// TODO: Require all runners to be deleted before deleting group or pass force=true parameter

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

func main() {
	//err := os.MkdirAll("logs", os.ModePerm)
	//if err != nil {
	//	panic(err)
	//}
	//f, err := os.OpenFile("logs/server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	//if err != nil {
	//	panic(err)
	//}
	//log.SetOutput(io.MultiWriter(os.Stdout, f))
	//
	config := &apis.Config{}
	bytes, err := ioutil.ReadFile("../config.yml")
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
	privateKey, err := base64.StdEncoding.DecodeString(config.PrivateKey)
	if err != nil {
		panic(err)
	}

	log.Info("Generating GitHub application credentials")
	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, config.InstallationID, privateKey)
	if err != nil {
		panic("Failed creating app authentication")
	}

	log.Info("Creating GitHub client")
	client := github.NewClient(&http.Client{Transport: itr})
	createClient := func(token string) (*apis.MaintainershipClient, error) {
		log.Info("Creating user GitHub client")
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		rateLimit, _, err := client.RateLimits(context.Background())
		if err != nil || rateLimit.GetCore().Limit != 5000 {
			return nil, fmt.Errorf("unable to verify authorization token authenticity: %w", err)
		}
		return &apis.MaintainershipClient{
			TeamsClient: client.Teams,
			UsersClient: client.Users,
		}, nil
	}
	manager := &apis.Manager{
		ActionsClient:              client.Actions,
		RepositoriesClient:         client.Repositories,
		TeamsClient:                client.Teams,
		Config:                     config,
		CreateMaintainershipClient: createClient,
	}

	http.HandleFunc("/group-create", manager.DoGroupCreate)
	http.HandleFunc("/group-delete", manager.DoGroupDelete)
	http.HandleFunc("/group-list", manager.DoGroupList)
	http.HandleFunc("/repos-add", manager.DoReposAdd)
	http.HandleFunc("/repos-remove", manager.DoReposRemove)
	http.HandleFunc("/repos-set", manager.DoReposSet)
	http.HandleFunc("/token-register", manager.DoTokenRegister)
	http.HandleFunc("/token-remove", manager.DoTokenRemove)

	err = http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}
