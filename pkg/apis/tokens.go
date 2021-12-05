package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func (m *Manager) DoTokenRegister(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"error": "User is not a maintainer of the team",
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
			"error": fmt.Sprintf("Unable to create registration token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner registration token")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": registrationToken,
	})
}

func (m *Manager) DoTokenRemove(c *gin.Context) {
	uuid := requestid.Get(c)

	m.Logger.Info("Retrieving team parameter")
	team := c.Query("team")
	if team == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":  http.StatusBadRequest,
			"error": "Missing required parameter: team",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved team parameter")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Retrieving Authorization header")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": "Missing Authorization header",
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Retrieved Authorization header")

	m.Logger.WithField("uuid", uuid).WithField("team", team).Info("Verifying maintainership")
	isMaintainer, err := m.verifyMaintainership(token, team, uuid)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"Code":  http.StatusForbidden,
			"error": fmt.Sprintf("Unable to validate user is a team maintainer: %v", err),
		})
		return
	}
	if !isMaintainer {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Code":  http.StatusUnauthorized,
			"error": "User is not a maintainer of the team",
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
			"error": fmt.Sprintf("Unable to create organization removal token: %v", err),
		})
		return
	}
	m.Logger.WithField("uuid", uuid).WithField("team", team).Debug("Created organization runner removal token")

	c.JSON(http.StatusOK, gin.H{
		"Code":     http.StatusOK,
		"Response": removalToken,
	})
}
