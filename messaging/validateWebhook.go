package messaging

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/client"
)

func ValidateWebhook(c *gin.Context) error {
	var params map[string]string

	requestValidator := client.NewRequestValidator(TwilioAuthToken)

	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	signature := c.GetHeader("X-Twilio-Signature")

	if !requestValidator.Validate(WebhookUrl, params, signature) {
		return errors.New("invalid signature")
	}

	return nil
}
