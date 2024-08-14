package messaging

import (
	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/client"
)

func ValidateWebhook(c *gin.Context) bool {
	var params map[string]string

	requestValidator := client.NewRequestValidator(TwilioAuthToken)

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return false
	}

	signature := c.GetHeader("X-Twilio-Signature")

	return requestValidator.Validate(WebhookUrl, params, signature)
}
