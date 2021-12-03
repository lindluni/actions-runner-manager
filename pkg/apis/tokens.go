package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *Manager) DoTokenRegister(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}

	ctx := context.Background()
	registrationToken, _, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create registration token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}

	bytes, err := json.Marshal(registrationToken)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, "Unable to marshal registration token")
		return
	}
	m.writeResponse(w, http.StatusOK, string(bytes))
}

func (m *Manager) DoTokenRemove(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: team")
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		m.writeResponse(w, http.StatusBadRequest, "Missing required parameter: token")
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		message := fmt.Sprintf("Unable to validate user is a team maintainer: %v", err)
		m.writeResponse(w, http.StatusForbidden, message)
		return
	}
	if !isMaintainer {
		m.writeResponse(w, http.StatusUnauthorized, "User is not a maintainer of the team")
		return
	}

	ctx := context.Background()
	removalToken, _, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		message := fmt.Sprintf("Unable to create removal token: %v", err)
		m.writeResponse(w, http.StatusInternalServerError, message)
		return
	}

	bytes, err := json.Marshal(removalToken)
	if err != nil {
		m.writeResponse(w, http.StatusInternalServerError, "Unable to marshal removal token")
		return
	}
	m.writeResponse(w, http.StatusOK, string(bytes))
}
