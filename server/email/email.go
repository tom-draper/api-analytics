package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func getEmailLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	address := os.Getenv("AUTOMATION_EMAIL_PASSWORD")
	password := os.Getenv("AUTOMATION_EMAIL_PASSWORD")
	return address, password
}

func GetEmailAddress() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	address := os.Getenv("EMAIL_ADDRESS")
	return address
}

// Login solution provided by andelf
// https://gist.github.com/andelf/5118732
type loginAuth struct {
	username string
	password string
}

func LoginAuth(username string, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

func SendEmail(subject string, body string, dest string) error {
	server := "smtp-mail.outlook.com"
	port := 587
	address, password := getEmailLogin()
	from := address

	fmt.Println(address, password)

	auth := LoginAuth(address, password)
	fmt.Println("logged in")

	to := []string{dest}

	msg := []byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\nOK", from, dest, subject))

	endpoint := fmt.Sprintf("%s:%d", server, port)
	fmt.Println(endpoint)
	err := smtp.SendMail(endpoint, auth, from, to, msg)
	return err
}
