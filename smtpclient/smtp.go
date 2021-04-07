package smtpclient

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/hokonco/kitgo"
)

type Config struct {
	Auth struct {
		Plain struct {
			Identity string `yaml:"identity" json:"identity"`
			Username string `yaml:"username" json:"username"`
			Password string `yaml:"password" json:"password"`
		} `yaml:"plain" json:"plain"`
		CRAMMD5 struct {
			Username string `yaml:"username" json:"username"`
			Secret   string `yaml:"secret" json:"secret"`
		} `yaml:"cram_md5" json:"cram_md5"`
	} `yaml:"auth" json:"auth"`
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

func New(cfg Config) *Client {
	auth := smtp.PlainAuth(cfg.Auth.Plain.Identity, cfg.Auth.Plain.Username, cfg.Auth.Plain.Password, cfg.Host)
	if cfg.Auth.CRAMMD5.Secret != "" && cfg.Auth.CRAMMD5.Username != "" {
		auth = smtp.CRAMMD5Auth(cfg.Auth.CRAMMD5.Username, cfg.Auth.CRAMMD5.Secret)
	}
	addr := cfg.Host
	if !strings.ContainsRune(cfg.Host, ':') && cfg.Port > 0 {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	return &Client{addr: addr, auth: auth}
}

func Test() (*Client, *Mock) {
	conf := Config{Host: "127.0.0.1", Port: 25}
	conf.Auth.CRAMMD5.Username = "username"
	conf.Auth.CRAMMD5.Secret = "secret"
	c := New(conf)
	return c, &Mock{c}
}

type Client struct {
	addr        string
	auth        smtp.Auth
	netResolver kitgo.NetResolver
	netDialer   kitgo.NetDialer

	newSmtpClient func(net.Conn, string) (kitgo.SmtpClient, error)

	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

func (x *Client) ReverseLookup(ctx context.Context, from, to mail.Address) (err error) {
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

	if x.newSmtpClient == nil {
		x.newSmtpClient = func(c net.Conn, h string) (kitgo.SmtpClient, error) { return smtp.NewClient(c, h) }
	}
	var client kitgo.SmtpClient
	if client, err = x.newSmtpClient(conn, host(to)); err != nil || client == nil {
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
func (x *Client) SendMail(ctx context.Context, id string, from, to mail.Address, subject string, plain, html []byte) error {
	id, subject = strings.TrimSpace(id), strings.TrimSpace(subject)
	if ok, err := validate(from, to); err != nil || !ok {
		return err
	} else if x.auth == nil || len(subject) < 1 || len(plain) < 1 || len(id) < 1 {
		return fmt.Errorf("smtp: required auth / subject / plain / message id")
	}
	if x.sendMail == nil {
		x.sendMail = smtp.SendMail
	}
	return x.sendMail(x.addr, x.auth, from.String(), []string{to.String()},
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

type Mock struct{ c *Client }

func (s *Mock) WithNetDialer(netDialer kitgo.NetDialer) {
	s.c.netDialer = netDialer
}
func (s *Mock) WithNetResolver(netResolver kitgo.NetResolver) {
	s.c.netResolver = netResolver
}
func (s *Mock) WithNewSmtpClient(client kitgo.SmtpClient, err error) {
	s.c.newSmtpClient = func(net.Conn, string) (kitgo.SmtpClient, error) { return client, err }
}
func (s *Mock) WithSendMail(err error) {
	s.c.sendMail = func(string, smtp.Auth, string, []string, []byte) error { return err }
}
