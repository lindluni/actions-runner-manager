package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// DoTokenRegister Request a GitHub Action organization runner registration token
// @Summary      Creates a new GitHub Action organization runner registration token
// @Description  Create a new GitHub Action organization runner registration token
// @Tags         tokens
// @Produce      json
// @Param        team   path      string  true  "Canonical **slug** of the GitHub team"
// @Success      200    {object}  github.RegistrationToken
// @Router       /token-register/{team} [get]
// @Security     ApiKeyAuth
func (m *Manager) DoTokenRegister(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"Error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"Error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"Error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"Error": "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Creating organization runner registration token")
	registrationToken, resp, err := m.ActionsClient.CreateOrganizationRegistrationToken(ctx, m.Config.Org)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{
			"Code":  resp.StatusCode,
			"Error": fmt.Sprintf("Unable to create registration token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner registration token")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": registrationToken,
	})
}

// DoTokenRemove Request a GitHub Action organization runner removal token
// @Summary      Creates a new GitHub Action organization runner removal token
// @Description  Create a new GitHub Action organization runner removal token
// @Tags         tokens
// @Produce      json
// @Param        team   path      string  true  "Canonical **slug** of the GitHub team"
// @Success      200    {object}  github.RegistrationToken
// @Router       /token-remove/{team} [get]
// @Security     ApiKeyAuth
func (m *Manager) DoTokenRemove(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"Error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"Error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"Error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"Error": "User is not a maintainer of the team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Verified maintainership")

	ctx := context.Background()
	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Creating organization runner removal token")
	removalToken, resp, err := m.ActionsClient.CreateOrganizationRemoveToken(ctx, m.Config.Org)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{
			"Code":  resp.StatusCode,
			"Error": fmt.Sprintf("Unable to create organization removal token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner removal token")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": removalToken,
	})
}
