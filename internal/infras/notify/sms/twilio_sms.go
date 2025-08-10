package sms

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TwilioSender struct {
	accountSID string
	authToken  string
	from       string
}

func NewTwilioSender(accountSID, authToken, from string) *TwilioSender {
	return &TwilioSender{accountSID: accountSID, authToken: authToken, from: from}
}

func (t *TwilioSender) Send(ctx context.Context, to, message string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", t.from)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio error: status=%d", resp.StatusCode)
	}
	return nil
}
