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
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		m.Logger.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		m.Logger.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	ctx := context.Background()
	registrationToken, _, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(registrationToken)
	if err != nil {
		m.Logger.Error(err)
	}
}

func (m *Manager) DoTokenRemove(w http.ResponseWriter, req *http.Request) {
	teamParam := req.URL.Query()["team"]
	if len(teamParam) != 1 {
		http.Error(w, "Missing required parameter: team", http.StatusBadRequest)
		return
	}
	team := teamParam[0]

	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "authorization header missing", http.StatusForbidden)
		return
	}

	isMaintainer, err := m.verifyMaintainership(token, team)
	if err != nil {
		m.Logger.Error(err)
		http.Error(w, fmt.Sprintf("Unable to validate user is a team maintainer: %v+", err), http.StatusForbidden)
		return
	}
	if !isMaintainer {
		m.Logger.Error("User is not a maintainer of the team")
		http.Error(w, "User is not a maintainer of the team", http.StatusForbidden)
		return
	}

	ctx := context.Background()
	removalToken, _, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(removalToken)
	if err != nil {
		m.Logger.Error(err)
	}
}
