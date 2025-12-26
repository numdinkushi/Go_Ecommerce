package notification

import (
	"encoding/json"
	"fmt"
	"go-ecommerce-app/config"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type NotificationClient interface {
	SendSMS(phone string, message string) error
}

type notificationClient struct {
	config config.AppConfig
}

// twillio
func (c notificationClient) SendSMS(phone string, message string) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: c.config.TwilioAccountSid,
		Password: c.config.TwilioAuthToken,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(phone)
	params.SetFrom(c.config.TwilioPhoneNumber)
	params.SetBody(message)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println("Error sending SMS message: " + err.Error())
		return err
	}

	response, _ := json.Marshal(*resp)
	fmt.Println("Response: " + string(response))
	return nil
}

func NewNotificationClient(config config.AppConfig) NotificationClient {
	return &notificationClient{config: config}
}
