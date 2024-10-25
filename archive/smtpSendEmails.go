package main

// Setup the following for the application
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/smtpClient"

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o smtpClient.bin -ldflags "-w -s" main.go

// To cross compile windows (Not tested...)
// GOOS=windows GOARCH=amd64 go build -o smtpClient.exe -ldflags "-w -s" main.go

Reference: https://linuxhint.com/golang-send-email/

// mailslurp...

*/

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

const (
	USERNAME = "linuxhint"
	PASSWD   = "password"
	HOST     = "mail.internal.test"
)

func main() {
	from := "dev@testing.io"
	to := []string{
		"dev.mail@mailsluper.com",
	}
	msg := []byte("Email Body: Welcome to Go!\r\n")
	err := SendMail(HOST+":2500", from, "Golang testing email", string(msg), to)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Mail Sent Successfully!")

}
func SendMail(addr, from, subject, body string, to []string) error {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "", "%0a", "", "%0d", "")

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Mail(r.Replace(from)); err != nil {
		return err
	}
	for i := range to {
		to[i] = r.Replace(to[i])
		if err = c.Rcpt(to[i]); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := "To: " + strings.Join(to, ",") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"X-Additional-Header: processbackdoor\r\n" +
		"\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
