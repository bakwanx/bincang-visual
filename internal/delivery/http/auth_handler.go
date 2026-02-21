package http

import (
	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/repository"
	"bincang-visual/internal/middleware"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type AuthHandler struct {
	userRepo     repository.UserRepository
	googleConfig *oauth2.Config
	jwtSecret    string
	jwtExpiry    int
	redisClient  *redis.Client
}

func NewAuthHandler(
	userRepo repository.UserRepository,
	googleConfig *oauth2.Config,
	jwtSecret string,
	jwtExpiry int,
	redisClient *redis.Client,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		googleConfig: googleConfig,
		jwtSecret:    jwtSecret,
		jwtExpiry:    jwtExpiry,
		redisClient:  redisClient,
	}
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	// Generate state token for CSRF protection
	state := generateStateToken()

	// Store state in cookie
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.JSON(fiber.Map{
		"url": url,
	})
}

func (h *AuthHandler) GoogleTokenSignIn(c *fiber.Ctx) error {
	var req struct {
		IDToken string `json:"idToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.IDToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "idToken is required",
		})
	}

	userInfo, err := h.verifyGoogleIDToken(req.IDToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	user, err := h.getOrCreateUser(context.Background(), userInfo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	jwtToken, err := middleware.GenerateJWT(
		user.ID,
		user.Email,
		user.DisplayName,
		h.jwtSecret,
		h.jwtExpiry,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token":     jwtToken,
		"user":      user,
		"expiresIn": h.jwtExpiry * 3600,
	})
}

func (h *AuthHandler) verifyGoogleIDToken(idToken string) (*GoogleUserInfo, error) {

	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token verification failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenInfo GoogleUserInfo
	if err := json.Unmarshal(data, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	// Verify the token is for your application
	// if tokenInfo.Aud != h.googleConfig.ClientID {
	// 	return nil, fmt.Errorf("token audience mismatch")
	// }

	return &tokenInfo, nil
}

func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {

	state := c.Query("state")
	savedState := c.Cookies("oauth_state")
	if state != savedState {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	code := c.Query("code")
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	userInfo, err := h.getUserInfo(token.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}

	user, err := h.getOrCreateUser(context.Background(), userInfo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	jwtToken, err := middleware.GenerateJWT(
		user.ID,
		user.Email,
		user.DisplayName,
		h.jwtSecret,
		h.jwtExpiry,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token":     jwtToken,
		"user":      user,
		"expiresIn": h.jwtExpiry * 3600,
	})
}

func (h *AuthHandler) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	email := c.Locals("email").(string)
	displayName := c.Locals("displayName").(string)

	token, err := middleware.GenerateJWT(
		userID,
		email,
		displayName,
		h.jwtSecret,
		h.jwtExpiry,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token":     token,
		"expiresIn": h.jwtExpiry * 3600,
	})
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := h.userRepo.Get(context.Background(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *AuthHandler) storeOAuthToken(userID string, token *oauth2.Token) error {
	key := fmt.Sprintf("oauth_token:%s", userID)
	tokenJSON, _ := json.Marshal(token)

	expiration := time.Until(token.Expiry)
	return h.redisClient.Set(context.Background(), key, tokenJSON, expiration).Err()
}

func (h *AuthHandler) getOrCreateUser(ctx context.Context, userInfo *GoogleUserInfo) (*entity.User, error) {

	user, err := h.userRepo.GetByEmail(ctx, userInfo.Email)
	if err == nil {
		// update info
		user.DisplayName = userInfo.Name
		user.PhotoURL = userInfo.Picture
		err = h.userRepo.Update(ctx, user)
		return user, err
	}

	user = &entity.User{
		ID:          userInfo.ID,
		Email:       userInfo.Email,
		DisplayName: userInfo.Name,
		PhotoURL:    userInfo.Picture,
		CreatedAt:   time.Now(),
	}

	err = h.userRepo.Create(ctx, user)
	return user, err
}

func generateStateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
