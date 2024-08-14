package messaging

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/client"
)

func ValidateWebhook(c *gin.Context) error {
	params := make(map[string]string)
	c.Request.ParseForm()
	for key, value := range c.Request.PostForm {
		params[key] = value[0]
	}

	requestValidator := client.NewRequestValidator(TwilioAuthToken)

	signature := c.Request.Header.Get("X-Twilio-Signature")

	if !requestValidator.Validate(TwilioWebhookUrl, params, signature) {
		return errors.New("invalid signature")
	}

	return nil
}
