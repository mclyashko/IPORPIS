package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/smtp"
	"os"
	"strings"
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
	// Указываем адреса получателя и отправителя
	if err := s.client.Mail(s.username); err != nil {
		return fmt.Errorf("error setting sender in SMTP client: %v", err)
	}
	if err := s.client.Rcpt(to); err != nil {
		return fmt.Errorf("error setting recipient in SMTP client: %v", err)
	}

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

	// Записываем сообщение
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("error writing data to SMTP writer: %v", err)
	}

	return nil
}

// Close закрывает SMTP-клиент
func (s *SMTPSender) Close() error {
	if quitErr := s.client.Quit(); quitErr != nil {
		return fmt.Errorf("error closing SMTP client: %v", quitErr)
	}
	return nil
}

// createMessage формирует MIME-сообщение
func (s *SMTPSender) createMessage(to, subject, body string, attachmentFilePaths []string) (string, error) {
	var msg strings.Builder
	writer := multipart.NewWriter(&msg)
	defer writer.Close() // Закрываем writer после завершения функции

	// Заголовки
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.username))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", writer.Boundary()))

	// Основное сообщение
	msg.WriteString(fmt.Sprintf("--%s\r\n", writer.Boundary()))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(body + "\r\n")

	// Прикрепление файлов
	for _, attachment := range attachmentFilePaths {
		if err := s.addFileAttachment(writer, attachment); err != nil {
			return "", err
		}
	}

	// Добавляем окончание сообщения
	msg.WriteString(fmt.Sprintf("--%s--\r\n", writer.Boundary()))

	return msg.String(), nil
}

// addFileAttachment добавляет файл как вложение в сообщение
func (s *SMTPSender) addFileAttachment(w *multipart.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening attachment file %s: %v", filename, err)
	}
	defer file.Close() // Закрываем файл после завершения функции

	// Создаем заголовок для вложения
	part, err := w.CreateFormFile("attachment", filename)
	if err != nil {
		return fmt.Errorf("error creating form file for attachment %s: %v", filename, err)
	}

	// Копируем содержимое файла в часть
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("error copying file content to part: %v", err)
	}

	return nil
}
