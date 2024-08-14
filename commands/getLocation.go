package commands

import (
	"github.com/Hodik/geo-tracker-be/messaging"
)

const (
	GetCurrentLocationCommand string = "999"
)

func SendGetDeviceLocationCommand(number string) (err error) {
	_, err = messaging.Send(string(number), GetCurrentLocationCommand)
	return err
}
