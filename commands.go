package main

import (
	"github.com/Hodik/geo-tracker-be/messaging"
)

const (
	GetCurrentLocationCommand string = "999"
	ResetCommand              string = "1122"
	RestartCommand            string = "SYSRST#"
)

func SendGetDeviceLocationCommand(number string) (err error) {
	_, err = messaging.Send(string(number), GetCurrentLocationCommand)
	return err
}

func SendResetCommand(number string) (err error) {
	_, err = messaging.Send(string(number), ResetCommand)
	return err
}

func SendRestartCommand(number string) (err error) {
	_, err = messaging.Send(string(number), RestartCommand)
	return err
}
