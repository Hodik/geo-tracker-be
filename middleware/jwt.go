package middleware

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

type CustomClaims struct {
	Email string `json:"email"`

	Scope         *string `json:"scope"`
	Name          *string `json:"name"`
	EmailVerified *bool   `json:"email_verified"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT in a Gin context.
func EnsureValidToken() gin.HandlerFunc {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator: %v", err)
	}

	errorHandler := func(c *gin.Context, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		c.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to validate JWT."})
		c.Abort()
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorHandler(c, err)
			return
		}

		// Extract the token from the Authorization header ("Bearer <token>")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			errorHandler(c, err)
			return
		}

		// Validate the token
		validatedClaims, err := jwtValidator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			errorHandler(c, err)
			return
		}

		// Add the validated token claims to the context for use in other handlers
		c.Set("token", validatedClaims)

		c.Next()
	}
}
