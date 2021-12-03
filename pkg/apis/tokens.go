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
	m.Logger.WithField("uuid", id).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}
	m.Logger.WithField("uuid", id).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).Info("Verifying maintainership")
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
	m.Logger.WithField("uuid", id).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).Info("Creating organization runner registration token")
	registrationToken, _, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create registration token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).Debug("Created organization runner registration token")

	//m.Logger.WithField("uuid", id).Info("Marshalling runner registration token")
	//bytes, err := json.Marshal(registrationToken)
	//if err != nil {
	//	m.writeResponse(w, http.StatusInternalServerError, "Unable to marshal registration token")
	//	return
	//}
	//m.Logger.WithField("uuid", id).Debug("Marshalled runner registration token")

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
	m.Logger.WithField("uuid", id).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", id).Info("Retrieving Authorization header")
	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}
	m.Logger.WithField("uuid", id).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", id).Info("Verifying maintainership")
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
	m.Logger.WithField("uuid", id).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", id).Info("Deleting organization runner registration token")
	removalToken, _, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create removal token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}
	m.Logger.WithField("uuid", id).Debug("Deleted organization runner registration token")

	//m.Logger.WithField("uuid", id).Info("Marshalling runner removal token")
	//bytes, err := json.Marshal(removalToken)
	//if err != nil {
	//	m.writeResponse(w, http.StatusInternalServerError, "Unable to marshal removal token")
	//	return
	//}
	//m.Logger.WithField("uuid", id).Debug("Marshalled runner removal token")

	response := &response{
		StatusCode: http.StatusOK,
		Response:   removalToken,
	}
	m.writeResponseWithUUID(w, response, id)
}
