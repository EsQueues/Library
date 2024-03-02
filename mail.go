package main

//
//import (
//	"fmt"
//	"net/smtp"
//)
//
//func sendMailSimple(subject string, body string, to []string) {
//	auth := smtp.PlainAuth(
//		"",
//		"saiat.kusainov05@gmail.com",
//		"mvip fblq yhtq gwqa",
//		"smtp.gmail.com")
//
//	msg := "Subject: " + subject + "\n" + body
//
//	err := smtp.SendMail(
//		"smtp.gmail.com:587",
//		auth,
//		"saiat.kusainov05@gmail.com",
//		to,
//		[]byte(msg),
//	)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//}
//
//func main() {
//	sendMailSimple("Second test", "Almassss ", []string{"nomad1465qs@gmail.com"})
//}
