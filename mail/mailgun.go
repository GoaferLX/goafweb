package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goafweb"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/mailgun/mailgun-go"
)

type mailService struct {
	mg mailgun.Mailgun
}
type ContactForm struct {
	Email   string `schema:"email"`
	Subject string `schema:"subject"`
	Message string `schema:"message"`
}

func NewMailService(domain, apiKey string) goafweb.MailService {
	mgclient := mailgun.NewMailgun(domain, apiKey)
	mgclient.SetAPIBase(mailgun.APIBaseEU)
	return &mailService{
		mg: mgclient,
	}
}
func (ms *mailService) Contact(w http.ResponseWriter, r *http.Request) {
	var form ContactForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
	if !emailRegex.MatchString(form.Email) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errors.New("Not a valid email"))
		return
	}
	filterRegex := regexp.MustCompile(`@leannesbowtique.com`)
	if filterRegex.MatchString(form.Email) {
		log.Print("Tried sending email from own domain")
		return
	}

	msg := ms.mg.NewMessage(form.Email, fmt.Sprint("LB Enquiry: "+form.Subject), form.Message, "leanne@leannesbowtique.com")
	msg.AddBCC("support@leannesbowtique.com")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err := ms.mg.Send(ctx, msg)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return

	}
	w.WriteHeader(http.StatusOK)
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
