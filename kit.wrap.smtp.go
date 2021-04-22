package kitgo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// SMTP implement a wrapper around standard "html/template"
// so that the SMTP is limited to Parse & Execute
var SMTP smtp_

type smtp_ struct{}

type SMTPConfig struct {
	PlainAuthIdentity   string
	PlainAuthUsername   string
	PlainAuthPassword   string
	CRAMMD5AuthUsername string
	CRAMMD5AuthSecret   string
	Addr                string
}

func (x smtp_) New(conf *SMTPConfig) *NetSMTPWrapper {
	PanicWhen(conf == nil || conf.Addr == "", conf)
	auth := smtp.PlainAuth(conf.PlainAuthIdentity, conf.PlainAuthUsername, conf.PlainAuthPassword, conf.Addr)
	if conf.CRAMMD5AuthUsername != "" && conf.CRAMMD5AuthSecret != "" {
		auth = smtp.CRAMMD5Auth(conf.CRAMMD5AuthUsername, conf.CRAMMD5AuthSecret)
	}
	return &NetSMTPWrapper{addr: conf.Addr, auth: auth}
}

func (smtp_) Test() (*NetSMTPWrapper, *NetSMTPMock) {
	c := SMTP.New(&SMTPConfig{CRAMMD5AuthUsername: "-", CRAMMD5AuthSecret: "-", Addr: "127.0.0.1:25"})
	return c, &NetSMTPMock{c}
}

type NetSMTPWrapper struct {
	addr        string
	auth        smtp.Auth
	netResolver NetResolver
	netDialer   NetDialer

	newFunc  func(net.Conn, string) (SmtpClient, error)
	sendFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func (x *NetSMTPWrapper) ReverseLookup(ctx context.Context, from, to mail.Address) (err error) {
	if ok, err := validate(from, to); err != nil || !ok {
		return err
	}
	var toMXs []*net.MX
	if toMXs, err = x.netResolver.LookupMX(ctx, host(to)); err != nil || len(toMXs) < 1 {
		return
	}
	var conn net.Conn
	if conn, err = x.netDialer.DialContext(ctx, "tcp", host(to)+":25"); err != nil || conn == nil {
		return
	}

	if x.newFunc == nil {
		x.newFunc = func(c net.Conn, h string) (SmtpClient, error) { return smtp.NewClient(c, h) }
	}
	var client SmtpClient
	if client, err = x.newFunc(conn, host(to)); err != nil || client == nil {
		return err
	}
	defer client.Close()
	if err = client.Hello(host(from)); err != nil {
		return
	}
	if err = client.Mail(from.Address); err != nil {
		return
	}
	err = client.Rcpt(to.Address)
	return
}
func (x *NetSMTPWrapper) SendMail(ctx context.Context, id string, from, to mail.Address, subject string, plain, html []byte) error {
	id, subject = strings.TrimSpace(id), strings.TrimSpace(subject)
	if ok, err := validate(from, to); err != nil || !ok {
		return err
	} else if x.auth == nil || len(subject) < 1 || len(plain) < 1 || len(id) < 1 {
		return fmt.Errorf("smtp: required auth / subject / plain / message id")
	}
	if x.sendFunc == nil {
		x.sendFunc = smtp.SendMail
	}
	return x.sendFunc(x.addr, x.auth, from.String(), []string{to.String()},
		payload(from, to, subject, time.Now(), id, plain, html),
	)
}
func host(a mail.Address) (host string) {
	if s := strings.Split(a.Address, "@"); len(s) > 1 && len(s[1]) > 0 {
		host = s[1]
	}
	return host
}
func validate(from, to mail.Address) (ok bool, err error) {
	validateMailAddress := func(a mail.Address) (ok bool, err error) {
		var m *mail.Address
		if m, err = mail.ParseAddress(a.String()); err == nil && m != nil && *m == a {
			ok = true
		}
		return ok, err
	}
	if from.Address == to.Address {
		return false, fmt.Errorf("smtp: same address")
	} else if ok, err := validateMailAddress(from); !ok {
		return false, fmt.Errorf("smtp: invalid from address (%w)", err)
	} else if ok, err := validateMailAddress(to); !ok {
		return false, fmt.Errorf("smtp: invalid to address (%w)", err)
	}
	return true, nil
}
func join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}
func payload(from, to mail.Address, subject string, date time.Time, id string, plain, html []byte) []byte {
	msg := join("\n",
		"MIME-version: 1.0",
		"From: %s",
		"To: %s",
		"Subject: %s",
		"Date: %s",
		"Message-Id: %s",
		"Content-Type: multipart/alternative; boundary=\"%s\"",
		"",
		"%s",
		"",
		"--%s--",
	)
	each := join("\n",
		"",
		"--%s",
		"Content-Type: text/%s; charset=\"utf-8\"",
		"Content-Transfer-Encoding: quoted-printable",
		"Content-Disposition: inline",
		"",
		"%s",
		"",
	)
	boundary := fmt.Sprintf("boundary-%s", id)
	body := []byte(fmt.Sprintf(each, boundary, "plain", plain))
	if len(html) > 0 {
		body = append(body, []byte(fmt.Sprintf(each, boundary, "html", html))...)
	}
	return []byte(fmt.Sprintf(msg,
		from.String(),
		to.String(),
		subject,
		time.Now().Format(http.TimeFormat),
		fmt.Sprintf("<mail-%s@%s>", id, host(from)),
		boundary,
		body,
		boundary,
	))
}

// ========================================
// MOCK
// ========================================

type NetSMTPMock struct{ c *NetSMTPWrapper }

func (x *NetSMTPMock) WithNetDialer(netDialer NetDialer) {
	x.c.netDialer = netDialer
}
func (x *NetSMTPMock) WithNetResolver(netResolver NetResolver) {
	x.c.netResolver = netResolver
}
func (x *NetSMTPMock) WithNewSmtpClient(client SmtpClient, err error) {
	x.c.newFunc = func(net.Conn, string) (SmtpClient, error) { return client, err }
}
func (x *NetSMTPMock) WithSendMail(err error) {
	x.c.sendFunc = func(string, smtp.Auth, string, []string, []byte) error { return err }
}
