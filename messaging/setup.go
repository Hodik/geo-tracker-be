package messaging

import (
	"os"

	"github.com/twilio/twilio-go"
)

var TwilioClient *twilio.RestClient
var TwilioPhoneNumber string
var TwilioAuthToken string
var TwilioWebhookUrl string

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

	TwilioWebhookUrl = os.Getenv("TWILIO_WEBHOOK_URL")

	if TwilioWebhookUrl == "" {
		panic("TWILIO_WEBHOOK_URL is required")
	}

}
