package messaging

import (
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

func Send(to string, msg string) (*string, error) {

	if TwilioClient == nil {
		panic("TwilioClient is not initialized")
	}

	params := &api.CreateMessageParams{}
	params.SetBody(msg)
	params.SetFrom(TwilioPhoneNumber)
	params.SetTo(to)

	resp, err := TwilioClient.Api.CreateMessage(params)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
