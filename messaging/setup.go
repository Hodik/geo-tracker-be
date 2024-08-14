package messaging

import (
	"os"

	"github.com/twilio/twilio-go"
)

var TwilioClient *twilio.RestClient
var TwilioPhoneNumber string
var TwilioAuthToken string
var WebhookUrl string

func Setup() {
	TwilioClient = twilio.NewRestClient()

	TwilioPhoneNumber = os.Getenv("TWILIO_PHONE_NUMBER")

	if TwilioPhoneNumber == "" {
		panic("TWILIO_PHONE_NUMBER is required")
	}

	TwilioAuthToken = os.Getenv("TWILIO_AUTH_TOKEN")

	if TwilioAuthToken == "" {
		panic("TWILIO_AUTH_TOKEN is required")
	}

	WebhookUrl = os.Getenv("WEBHOOK_URL")

	if WebhookUrl == "" {
		panic("WEBHOOK_URL is required")
	}

}
