package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/exp/rand"

	"github.com/mclyashko/IPORPIS/internal/config"
	"github.com/mclyashko/IPORPIS/internal/email"
)

// emailRequest представляет тело запроса для отправки письма
type emailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// mailHandler обрабатывает запросы на отправку письма
func mailHandler(w http.ResponseWriter, r *http.Request, es email.Sender) {
	var emailReq emailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailReq); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Вызываем функцию для отправки письма
	if err := es.Send(emailReq.To, emailReq.Subject, emailReq.Body, []string{}); err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Ошибка отправки письма", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Письмо успешно отправлено")
}

func getConfig() config.App {
	configLoader := &config.DotenvConfigLoader{}

	appConfig, err := configLoader.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return appConfig
}

func main() {
	rand.Seed(uint64(time.Now().UnixNano()))

	cfg := getConfig()
	es, err := email.NewSMTPSender(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
	)
	if err != nil {
		log.Fatalf("Cant get SMTP sender: %v", err)
	}

	http.HandleFunc("POST /mail", func(w http.ResponseWriter, r *http.Request) {
		mailHandler(w, r, es)
	})

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
