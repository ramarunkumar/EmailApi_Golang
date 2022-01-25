package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"crypto/tls"

	"github.com/gin-gonic/gin"
	"github.com/k3a/html2text"
	_ "github.com/lib/pq"
	gomail "gopkg.in/mail.v2"
)

func main() {
	router := gin.Default()
	router.POST("/mail", mail)
	router.Run()
}

func mail(c *gin.Context) {

	db := dbinit()
	var emp Email

	if err := c.ShouldBindJSON(&emp); err != nil {
		fmt.Println("error in bind", err)
		return
	}

	mail := gomail.NewMessage()

	mail.SetHeader("From", "email@gmail.com")

	mail.SetHeader("To", emp.TO)

	mail.SetHeader("Subject", "Leave Required from "+emp.Startdate+" to "+emp.Enddate)

	stmt := "SELECT * FROM templates where id='" + emp.Id + "'"

	fmt.Println(stmt)

	err := db.QueryRow(stmt).Scan(&emp.Id, &emp.Title, &emp.Templates)
	if err != nil {
		fmt.Println("error", err)
	}

	html := emp.Templates

	plain := html2text.HTML2Text(html)

	f, err := os.Create("temp.txt")
	if err != nil {
		log.Fatal(err)
	}
	_, err2 := f.WriteString(plain)
	if err2 != nil {
		log.Fatal(err2)
	}
	err = ioutil.WriteFile("temp.html", []byte(html), 0777)
	if err != nil {
		fmt.Println(err)
	}
	data, err := ioutil.ReadFile("temp.txt")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(data))
	t, err := template.ParseFiles("temp.html")
	if err != nil {
		fmt.Println("Error happend", err)
		return
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, emp); err != nil {
		fmt.Println("error", err)
	}
	body := buf.String()

	mail.SetBody("text/html", body)

	send := gomail.NewDialer("smtp.gmail.com", 587, "email@gmail.com", "Password")

	send.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := send.DialAndSend(mail); err != nil {
		fmt.Println("error", err)
		c.IndentedJSON(http.StatusOK, gin.H{
			"Message": "Message sent failed",
		})
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"Message": "Message Sent sucessfull",
		"data":    emp,
	})
	fmt.Println("Message sent successfully")
}

type Email struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	TO        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Templates string `json:"templates"`
	Startdate string `json:"startdate"`
	Enddate   string `json:"Enddate"`
}

func dbinit() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:qwerty123@localhost:5432/email")
	if err != nil {
		fmt.Println("could not connect to database: ", err)
	}
	return db
}
