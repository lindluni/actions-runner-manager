package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (m *Manager) DoTokenRegister(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, id)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Creating organization runner registration token")
	registrationToken, _, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create registration token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Created organization runner registration token")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   registrationToken,
	}
	m.writeResponseWithUUID(w, response, id)
}

func (m *Manager) DoTokenRemove(w http.ResponseWriter, req *http.Request) {
	id := uuid.New().String()
	m.Logger.WithField("uuid", id).Info("Retrieving team parameter")
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, id)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).WithField("team", team).Info("Creating organization runner removal token")
	removalToken, _, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create removal token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).WithField("team", team).Debug("Create organization runner removal token")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   removalToken,
	}
	m.writeResponseWithUUID(w, response, id)
}
