package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
)

// Sender определяет интерфейс для отправки электронной почты
type Sender interface {
	Send(to, subject, body string, attachmentFilePaths []string) error
	Close() error
}

// SMTPSender реализует интерфейс Sender и отправляет почту через SMTP
type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	client   *smtp.Client
}

// NewSmtpEmailSender создает новый экземпляр SmtpEmailSender
func NewSMTPSender(host, port, username, password string) (*SMTPSender, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to SMTP server: %v", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, fmt.Errorf("error creating SMTP client: %v", err)
	}

	auth := smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		return nil, fmt.Errorf("error authenticating SMTP: %v", err)
	}

	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		client:   client,
	}, nil
}

// Send отправляет электронное письмо
func (s *SMTPSender) Send(to, subject, body string, attachmentFilePaths []string) error {
	log.Println("Начинаем отправку письма...")

	// Проверяем соединение
	if err := s.client.Noop(); err != nil {
		return fmt.Errorf("error checking SMTP connection: %v", err)
	}
	log.Println("Соединение с SMTP активно")

	// Указываем отправителя
	if err := s.client.Mail(s.username); err != nil {
		return fmt.Errorf("error setting sender in SMTP client: %v", err)
	}
	log.Println("Отправитель установлен:", s.username)

	// Указываем получателя
	if err := s.client.Rcpt(to); err != nil {
		return fmt.Errorf("error setting recipient in SMTP client: %v", err)
	}
	log.Println("Получатель установлен:", to)

	// Получаем writer для сообщения
	w, err := s.client.Data()
	if err != nil {
		return fmt.Errorf("error getting SMTP writer: %v", err)
	}
	defer w.Close() // Закрываем writer после завершения функции

	// Формируем сообщение
	message, err := s.createMessage(to, subject, body, attachmentFilePaths)
	if err != nil {
		return err
	}
	log.Println("Сообщение создано")

	// Записываем сообщение
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("error writing data to SMTP writer: %v", err)
	}

	log.Println("Письмо отправлено!")
	return nil
}

// Close закрывает SMTP-клиент
func (s *SMTPSender) Close() error {
	if quitErr := s.client.Quit(); quitErr != nil {
		return fmt.Errorf("error closing SMTP client: %v", quitErr)
	}
	return nil
}

func (s *SMTPSender) createMessage(to, subject, body string, attachmentFilePaths []string) (string, error) {
	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)

	// Заголовки письма
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n", 
		s.username, to, subject, writer.Boundary(),
	)
	msg.WriteString(headers)

	// Основное тело письма
	if err := addTextPart(writer, body); err != nil {
		return "", err
	}

	// Вложения
	for _, attachment := range attachmentFilePaths {
		if err := addFileAttachment(writer, attachment); err != nil {
			return "", err
		}
	}

	// Завершаем сообщение
	writer.Close()

	return msg.String(), nil
}

func addTextPart(w *multipart.Writer, body string) error {
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=UTF-8"},
		"Content-Transfer-Encoding": {"7bit"},
	})
	if err != nil {
		return fmt.Errorf("error creating text part: %v", err)
	}
	_, err = part.Write([]byte(body))
	return err
}

func addFileAttachment(w *multipart.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening attachment file %s: %v", filename, err)
	}
	defer file.Close()

	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Создаем заголовки для вложения
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filename))},
		"Content-Type":              {fmt.Sprintf("%s; name=\"%s\"", mimeType, filepath.Base(filename))},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return fmt.Errorf("error creating attachment part: %v", err)
	}

	// Кодируем файл в base64
	encoder := base64.NewEncoder(base64.StdEncoding, part)
	defer encoder.Close()

	if _, err = io.Copy(encoder, file); err != nil {
		return fmt.Errorf("error encoding file content: %v", err)
	}

	return nil
}