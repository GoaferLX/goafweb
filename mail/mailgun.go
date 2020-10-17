/*
Package mail handles sending mail using third party mailgun.
*/
package mail

import (
	"context"
	"fmt"
	"goafweb"
	"net/url"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailService struct {
	mg mailgun.Mailgun
}

// NewMailService returns a service implementing mailgun that fulfils
// goafweb.MailService interface.
func NewMailService(domain, apiKey string) goafweb.MailService {
	mgclient := mailgun.NewMailgun(domain, apiKey)
	mgclient.SetAPIBase(mailgun.APIBaseEU)
	return &mailService{
		mg: mgclient,
	}
}

const (
	resetPWSubject = "Instructions for resetting your password."
)
const resetTextTmpl = `Hi there!

It appears that you have requested a password reset. If this was you, please follow the link below to update your password:

%s

If you are asked for a token, please use the following value:

%s

If you didn't request a password reset you can safely ignore this email and your account will not be changed.

All the best,
Leanne @ Leanne's Bowtique`

const resetHTMLTmpl = `Hi there!<br/>
<br/>
It appears that you have requested a password reset. If this was you, please follow the link below to update your password:<br/>
<br/>
<a href="%s">%s</a><br/>
<br/>
If you are asked for a token, please use the following value:<br/>
<br/>
%s<br/>
<br/>
If you didn't request a password reset you can safely ignore this email and your account will not be changed.<br/>
<br/>
All the best,<br />
Leanne @ Leanne's Bowtique`

// ResetPW sends a reset token to the user provided email address.
func (ms *mailService) ResetPw(toEmail, token string) error {
	v := url.Values{}
	v.Set("token", token)
	resetURL := "https://leannesbowtique.com/reset?" + v.Encode()
	resetText := fmt.Sprintf(resetTextTmpl, resetURL, token)

	message := ms.mg.NewMessage("Leanne <support@leannesbowtique.com>", resetPWSubject, resetText, toEmail)
	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetURL, resetURL, token)
	message.SetHtml(resetHTML)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err := ms.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun Error, could not send: %w", err)
	}
	return nil
}
