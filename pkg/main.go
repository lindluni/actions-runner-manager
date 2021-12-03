/**
SPDX-License-Identifier: Apache-2.0
*/

// TODO: Allow user to override config path with env variable
// TODO: Implement pagination for github calls
// TODO: Reimplement GETS as POSTS, this will require creating structs to marshal the body into
// TODO: Add CODEOWNERS and enforce it
// TODO: Implement rate limits on user
// TODO: Require all runners to be deleted or ?transferred? before deleting group or pass force=true parameter

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"
	"github.com/lindluni/actions-runner-manager/pkg/apis"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

func main() {
	config, privateKey := initConfig()
	logger := initLogger(config)

	logger.Debug("Creating GitHub application installation configuration")
	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, config.InstallationID, privateKey)
	if err != nil {
		logger.Fatalf("Failed creating app authentication: %v", err)
	}
	logger.Debug("Created GitHub application installation configuration")

	logger.Debug("Creating GitHub client")
	client := github.NewClient(&http.Client{Transport: itr})
	logger.Debug("Created GitHub client")

	logger.Debug("Creating GitHub user client function")
	createClient := func(token, uuid string) (*apis.MaintainershipClient, error) {
		logger.WithField("uuid", uuid).Info("Creating GitHub user client")
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		logger.WithField("uuid", uuid).Debug("Created GitHub user client")

		logger.WithField("uuid", uuid).Info("Validating Authorization token")
		rateLimit, _, err := client.RateLimits(context.Background())
		if err != nil || rateLimit.GetCore().Limit != 5000 {
			logger.Errorf("Unable to verify authorization token authenticity: %v", err)
			return nil, fmt.Errorf("unable to verify authorization token authenticity: %w", err)
		}
		logger.WithField("uuid", uuid).Debug("Validated Authorization token")
		return &apis.MaintainershipClient{
			TeamsClient: client.Teams,
			UsersClient: client.Users,
		}, nil
	}

	logger.Debug("Creating API manager")
	manager := &apis.Manager{
		ActionsClient:              client.Actions,
		RepositoriesClient:         client.Repositories,
		TeamsClient:                client.Teams,
		Config:                     config,
		Logger:                     logger,
		CreateMaintainershipClient: createClient,
	}
	logger.Debug("Created API manager")

	logger.Info("Initializing API endpoints")
	http.HandleFunc("/group-create", manager.DoGroupCreate)
	http.HandleFunc("/group-delete", manager.DoGroupDelete)
	http.HandleFunc("/group-list", manager.DoGroupList)
	http.HandleFunc("/repos-add", manager.DoReposAdd)
	http.HandleFunc("/repos-remove", manager.DoReposRemove)
	http.HandleFunc("/repos-set", manager.DoReposSet)
	http.HandleFunc("/token-register", manager.DoTokenRegister)
	http.HandleFunc("/token-remove", manager.DoTokenRemove)
	logger.Debug("Initialized API endpoints")

	logger.Debug("Compiling HTTP server address")
	address := fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port)
	logger.Infof("Starting API server on address: %s", address)
	if config.Server.TLS.Enabled {
		err = http.ListenAndServeTLS(address, config.Server.TLS.CertFile, config.Server.TLS.KeyFile, nil)
		if err != nil {
			logger.Fatalf("API server failed: %v", err)
		}
	} else {
		err = http.ListenAndServe(address, nil)
		if err != nil {
			logger.Fatalf("API server failed: %v", err)
		}
	}
}

func initConfig() (*apis.Config, []byte) {
	var bytes []byte
	var err error
	logrus.Info("Loading configuration")
	configPath, set := os.LookupEnv("CONFIG_PATH")
	if set {
		bytes, err = ioutil.ReadFile(configPath)
	} else {
		bytes, err = ioutil.ReadFile("config.yml")
	}
	if err != nil {
		logrus.Fatalf("Unable to parse config file: %v", err)
	}
	logrus.Info("Configuration loaded")

	logrus.Info("Parsing configuration")
	config := &apis.Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		logrus.Fatalf("Unable to parse config file: %v", err)
	}
	logrus.Info("Configuration parsed")

	logrus.Info("Validating configuration")
	if !config.Logging.Ephemeral {
		if config.Logging.LogDirectory == "" || config.Logging.MaxSize <= 0 || config.Logging.MaxAge <= 0 {
			logrus.Fatal("Logging in non-ephemeral mode requires you set the following logging values: logDirectory, maxAge, maxSize")
		}
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	logrus.Info("Configuration validated")

	logrus.Info("Decoding private key")
	privateKey, err := base64.StdEncoding.DecodeString(config.PrivateKey)
	if err != nil {
		logrus.Fatalf("Unable to decode private key from base64: %v", err)
	}
	logrus.Info("Private key decoded")
	return config, privateKey
}

func initLogger(config *apis.Config) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		logrus.Fatalf("Unable to parse logging level: %v", err)
	}
	logger.SetLevel(level)

	logger.Debug("Marshalling logging configuration")
	bytes, err := json.Marshal(config.Logging)
	if err != nil {
		logger.Fatalf("Unable to marshal logging configuration: %v", err)
	}
	logger.Debug("Marshalled logging configuration")

	logger.Debugf("Initializing logger with configuration: %s", string(bytes))
	if !config.Logging.Ephemeral {
		logPath := filepath.Join(config.Logging.LogDirectory, "/actions-runner-manager/server.log")
		logger.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Compress:   config.Logging.Compression,
			Filename:   logPath,
			MaxBackups: config.Logging.MaxBackups,
			MaxAge:     config.Logging.MaxAge,
			MaxSize:    config.Logging.MaxSize,
		}))
	}
	logger.Debug("Logger initialized")
	return logger
}
