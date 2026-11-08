package email

import (
	"context"
	"errors"
	"fmt"
	stdmail "net/mail"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Security string
	Timeout  time.Duration
}

type PasswordResetSender struct {
	client *gomail.Client
	from   string
}

func NewPasswordResetSender(cfg SMTPConfig) (*PasswordResetSender, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("SMTP host is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("invalid SMTP port")
	}
	if _, err := stdmail.ParseAddress(cfg.From); err != nil {
		return nil, fmt.Errorf("invalid SMTP from address: %w", err)
	}
	if (cfg.Username == "") != (cfg.Password == "") {
		return nil, errors.New(
			"SMTP username and password must be set together",
		)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	options := []gomail.Option{
		gomail.WithPort(cfg.Port),
		gomail.WithTimeout(cfg.Timeout),
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Security)) {
	case "none":
		options = append(
			options,
			gomail.WithTLSPolicy(gomail.NoTLS),
		)

	case "starttls":
		options = append(
			options,
			gomail.WithTLSPolicy(gomail.TLSMandatory),
		)

	case "tls":
		options = append(options, gomail.WithSSL())

	default:
		return nil, errors.New(
			"SMTP security must be one of: none, starttls, tls",
		)
	}

	if cfg.Username != "" {
		options = append(
			options,
			gomail.WithUsername(cfg.Username),
			gomail.WithPassword(cfg.Password),
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		)
	}

	client, err := gomail.NewClient(cfg.Host, options...)
	if err != nil {
		return nil, fmt.Errorf("create SMTP client: %w", err)
	}

	return &PasswordResetSender{
		client: client,
		from:   cfg.From,
	}, nil
}

func (s *PasswordResetSender) SendPasswordResetCode(ctx context.Context, to string, code string) error {
	message := gomail.NewMsg()
	if err := message.From(s.from); err != nil {
		return err
	}
	if err := message.To(to); err != nil {
		return err
	}
	message.Subject("Восстановление пароля Wagecalc")
	message.SetDate()
	message.SetMessageID()
	message.SetBodyString(
		gomail.TypeTextPlain,
		fmt.Sprintf(
			"Код восстановления пароля: %s\n\n"+
				"Код действует 15 минут.\n"+
				"Если вы не запрашивали восстановление, "+
				"проигнорируйте это письмо.\n",
			code,
		),
	)
	if err := s.client.DialAndSendWithContext(
		ctx,
		message,
	); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}
