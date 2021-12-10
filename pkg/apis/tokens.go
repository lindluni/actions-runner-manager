/**
SPDX-License-Identifier: Apache-2.0
*/

package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// DoTokenRegister Request a GitHub Action organization runner registration token
// @Summary      Create a new GitHub Action organization runner registration token
// @Description  Creates a new GitHub Action organization runner removal token that can be used to configure GitHub Action runners at the organization level
// @Tags         Tokens
// @Produce      json
// @Param        team  query     string  true  "Canonical **slug** of the GitHub team"
// @Success      200   {object}  JSONResultSuccess{Code=int,Response=github.RegistrationToken}
// @Router       /token-register [get]
// @Security     ApiKeyAuth
func (m *Manager) DoTokenRegister(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, &JSONResultError{
			Code:  http.StatusUnauthorized,
			Error: "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Creating organization runner registration token")
	registrationToken, resp, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		c.JSON(resp.StatusCode, &JSONResultError{
			Code:  resp.StatusCode,
			Error: fmt.Sprintf("Unable to create registration token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner registration token")

	c.JSON(http.StatusOK, &JSONResultSuccess{
		Code:     http.StatusOK,
		Response: registrationToken,
	})
}

// DoTokenRemove Request a GitHub Action organization runner removal token
// @Summary      Create a new GitHub Action organization runner removal token
// @Description  Creates a new GitHub Action organization runner removal token that can be used remove a GitHub Action runners at the organization level
// @Tags         Tokens
// @Produce      json
// @Param        team  query     string  true  "Canonical **slug** of the GitHub team"
// @Success      200   {object}  JSONResultSuccess{Code=int,Response=github.RegistrationToken}
// @Router       /token-remove [get]
// @Security     ApiKeyAuth
func (m *Manager) DoTokenRemove(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, &JSONResultError{
			Code:  http.StatusBadRequest,
			Error: "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, &JSONResultError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, &JSONResultError{
			Code:  http.StatusUnauthorized,
			Error: "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Creating organization runner removal token")
	removalToken, resp, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		c.JSON(resp.StatusCode, &JSONResultError{
			Code:  resp.StatusCode,
			Error: fmt.Sprintf("Unable to create organization removal token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner removal token")

	c.JSON(http.StatusOK, &JSONResultSuccess{
		Code:     http.StatusOK,
		Response: removalToken,
	})
}
