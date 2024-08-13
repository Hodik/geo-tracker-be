package messaging

import (
	"os"

	"github.com/twilio/twilio-go"
)

var TwilioClient *twilio.RestClient
var TwilioPhoneNumber string

func Setup() {
	TwilioClient = twilio.NewRestClient()

	TwilioPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")

	if TwilioPhoneNumber == "" {
		panic("TWILIO_PHONE_NUMBER is required")
	}
}
