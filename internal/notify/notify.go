package notify

import (
	"errors"
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Config struct {
	AccountSID string `json:"accountSID"`
	AuthToken  string `json:"authToken"`
}

func DefaultConfig() Config {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")

	if accountSID != "" && authToken != "" {
		return Config{
			AccountSID: accountSID,
			AuthToken:  authToken,
		}
	}

	return Config{}
}

type Message struct {
	From string `json:"from"`
	To   string `json:"to"`

	Body string `json:"body"`
}

func Send(conf Config, msg Message) error {
	if msg.From == "" || msg.To == "" {
		return errors.New("missing SMS from/to")
	}
	if msg.Body == "" {
		return nil
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: conf.AccountSID,
		Password: conf.AuthToken,
	})
	if client == nil {
		return errors.New("nil twilio client created")
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(msg.To)
	params.SetFrom(msg.From)
	params.SetBody(msg.Body)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("problem sending sms message: %v", err)
	}
	if resp != nil && resp.ErrorMessage != nil {
		return fmt.Errorf("sending sms failed: %v", *resp.ErrorMessage)
	}
	return nil
}
