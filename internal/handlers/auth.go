package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/reduan2660/swapenv-server/internal/auth"
	"github.com/reduan2660/swapenv-server/internal/models"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB                 *gorm.DB
	GithubClientID     string
	GithubClientSecret string
	JWTSecret          string
}

type PollRequest struct {
	DeviceCode string `json:"device_code"`
}

// POST /auth/device
func (h *AuthHandler) DeviceCode(c echo.Context) error {
	resp, err := auth.RequestDeviceCode(h.GithubClientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to request device code"})
	}

	return c.JSON(http.StatusOK, resp)
}

// POST /auth/poll
func (h *AuthHandler) Poll(c echo.Context) error {
	var req PollRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	tokenResp, err := auth.PollAccessToken(h.GithubClientID, h.GithubClientSecret, req.DeviceCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to poll token"})
	}

	if tokenResp.Error == "authorization_pending" {
		return c.JSON(http.StatusAccepted, map[string]string{"status": "pending"})
	}

	if tokenResp.Error == "slow_down" {
		return c.JSON(http.StatusAccepted, map[string]string{"status": "slow_down"})
	}

	if tokenResp.Error != "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": tokenResp.Error})
	}

	ghUser, err := auth.GetGithubUser(tokenResp.AccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
	}

	user, err := h.findOrCreateUser(ghUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	token, err := auth.GenerateToken(user.ID, user.OrgID, h.JWTSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token":   token,
		"user_id": user.ID,
		"org_id":  user.OrgID,
	})
}

func (h *AuthHandler) findOrCreateUser(ghUser *auth.GithubUser) (*models.User, error) {
	var user models.User
	githubId := fmt.Sprintf("%d", ghUser.ID)

	err := h.DB.Where("github_id = ?", githubId).First(&user).Error
	if err == nil {
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// TODO: re-think this when implementing team plan
	org := models.Organization{
		ID:        uuid.New(),
		Name:      ghUser.Login + "'s org",
		CloudSync: false,
		Type:      "personal",
	}

	if err := h.DB.Create(&org).Error; err != nil {
		return nil, err
	}

	// create user
	var email *string
	if ghUser.Email != "" {
		email = &ghUser.Email
	}

	var name *string
	if ghUser.Name != "" {
		name = &ghUser.Name
	}

	user = models.User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		GithubID:  githubId,
		OrgID:     org.ID,
		Role:      "admin",
		Can_share: true,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
