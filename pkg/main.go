/**
SPDX-License-Identifier: Apache-2.0
*/

// TODO: Ensure all http error response paths return at the end
// TODO: Move secrets to environment to protect them and add approvers for PR's

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
	"github.com/google/uuid"
	"github.com/lindluni/actions-runner-manager/pkg/apis"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"golang.org/x/oauth2"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	config, privateKey := initConfig()
	logger := initLogger(config)

	logger.Debug("Creating GitHub application installation configuration")
	itr, err := ghinstallation.New(http.DefaultTransport, config.AppID, config.InstallationID, privateKey)
	if err != nil {
		logger.Fatalf("Failed creating app authentication: %v", err)
	}
	logger.Debug("Created GitHub application installation configuration")

	logger.Info("Initializing Rate Limiter")
	lmt := tollbooth.NewLimiter(5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetHeader("Authorization", []string{})
	lmt.SetHeaderEntryExpirationTTL(time.Hour)
	lmt.SetMessage(`{"code":429,"response":"You have reached maximum request limit. Please try again in a few seconds."}`)
	lmt.SetMessageContentType("application/json")
	logger.Debug("Initialized Rate Limiter")

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
		user, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			logger.WithField("uuid", uuid).Errorf("Unable to verify authorization token authenticity: %v", err)
			return nil, fmt.Errorf("unable to verify authorization token authenticity: %w", err)
		}
		lmt.SetBasicAuthUsers(append(lmt.GetBasicAuthUsers(), user.GetLogin()))
		logger.WithField("uuid", uuid).Debug("Validated Authorization token")
		return &apis.MaintainershipClient{
			TeamsClient: client.Teams,
			UsersClient: client.Users,
		}, nil
	}

	logger.Debug("Creating GitHub client")
	client := github.NewClient(&http.Client{Transport: itr})
	logger.Debug("Created GitHub client")

	logger.Info("Initialize Router")
	router := gin.New()
	router.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	router.Use(gin.Logger())
	logger.Debug("Initialized Router")

	logger.Debug("Creating API manager")
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
	logger.Debug("Created API manager")

	manager.Serve()
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
		rotator := &lumberjack.Logger{
			Compress:   config.Logging.Compression,
			Filename:   logPath,
			MaxBackups: config.Logging.MaxBackups,
			MaxAge:     config.Logging.MaxAge,
			MaxSize:    config.Logging.MaxSize,
		}
		logger.SetOutput(ioutil.Discard)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logger.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
			Writer: io.MultiWriter(os.Stderr, rotator),
			LogLevels: []logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
			},
		})
		logger.AddHook(&writer.Hook{ // Send info and d	ebug logs to stdout
			Writer: io.MultiWriter(os.Stdout, rotator),
			LogLevels: []logrus.Level{
				logrus.InfoLevel,
				logrus.DebugLevel,
			},
		})

	}
	logger.Debug("Logger initialized")
	return logger
}
