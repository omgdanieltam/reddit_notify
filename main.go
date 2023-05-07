package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"gopkg.in/ini.v1"
	gomail "gopkg.in/mail.v2"
)

var (
	subreddit, keyword, smtpServer, smtpTo, smtpFrom, smtpUsername, smtpPassword string
	interval, smtpPort                                                           int
)

func main() {
	// Load config
	cfg, err := ini.Load("config.ini")

	if err != nil {
		quitConfigParseError(err.Error())
	}

	// Parse and check config values
	if cfg.Section("app").HasKey("subreddit") {
		subreddit = cfg.Section("app").Key("subreddit").String()
	} else {
		quitConfigParseError("Missing 'subreddit'")
	}

	if cfg.Section("app").HasKey("interval") {
		// default to 5 minutes
		interval = cfg.Section("app").Key("interval").MustInt(5)
	} else {
		quitConfigParseError("Missing 'interval'")
	}

	if cfg.Section("app").HasKey("keyword") {
		keyword = cfg.Section("app").Key("keyword").String()
	} else {
		quitConfigParseError("Missing 'keyword'")
	}

	if cfg.Section("smtp").HasKey("smtp_server") {
		smtpServer = cfg.Section("smtp").Key("smtp_server").String()
	} else {
		quitConfigParseError("Missing 'smtp_server'")
	}

	if cfg.Section("smtp").HasKey("smtp_port") {
		// default to port 25
		smtpPort = cfg.Section("smtp").Key("smtp_port").MustInt(25)
	} else {
		quitConfigParseError("Missing 'smtp_port'")
	}

	if cfg.Section("smtp").HasKey("smtp_username") {
		smtpUsername = cfg.Section("smtp").Key("smtp_username").String()
	} else {
		quitConfigParseError("Missing 'smtp_username'")
	}

	if cfg.Section("smtp").HasKey("smtp_password") {
		smtpPassword = cfg.Section("smtp").Key("smtp_password").String()
	} else {
		quitConfigParseError("Missing 'smtp_password'")
	}

	if cfg.Section("smtp").HasKey("smtp_to") {
		smtpTo = cfg.Section("smtp").Key("smtp_to").String()
	} else {
		quitConfigParseError("Missing 'smtp_to'")
	}

	if cfg.Section("smtp").HasKey("smtp_from") {
		smtpFrom = cfg.Section("smtp").Key("smtp_from").String()
	} else {
		quitConfigParseError("Missing 'smtp_from'")
	}

	loop()

}

func loop() {
	// Setup
	subreddit_rss := "https://www.reddit.com/r/" + subreddit + "/new/.json"

	client := &http.Client{}
	req, err := http.NewRequest("GET", subreddit_rss, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Set a header otherwise it'll be blocked
	req.Header.Set("User-Agent", "Golang_Reddit_Notif/1.0")

	// Continually GET subreddit for interval time
	for {

		// Get from subreddit
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		// Response from subreddit
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Amount of items
		limit64, _ := jsonparser.GetInt(body, "data", "dist")
		limit := int(limit64)

		// Loop through titles and texts
		for i := 0; i < limit; i++ {
			index := strconv.Itoa(i)
			title, _ := jsonparser.GetString(body, "data", "children", "["+index+"]", "data", "title")
			text, _ := jsonparser.GetString(body, "data", "children", "["+index+"]", "data", "selftext")
			alert := false

			// Check if keyword matches
			if strings.Contains(strings.ToLower(title), keyword) {
				alert = true
			} else if strings.Contains(strings.ToLower(text), keyword) {
				alert = true
			}

			// Send alert if keyword found
			if alert {
				url, _ := jsonparser.GetString(body, "data", "children", "["+index+"]", "data", "url")
				timestamp, _ := jsonparser.GetFloat(body, "data", "children", "["+index+"]", "data", "created_utc")
				validateAlert(title, text, url, int64(timestamp))
			}

		}

		// Sleep for interval time
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func quitConfigParseError(msg string) {
	fmt.Println("Error parsing config.ini: ", msg)
	os.Exit(1)
}

// Validate the alert to ensure that it needs to be sent
func validateAlert(title string, text string, url string, timestamp int64) {
	// Get timestamp of interval period
	currentTs := time.Now()
	intervalTs := currentTs.Add(-time.Minute * time.Duration(interval))

	// Only send alert if it's newer than interval time period
	if timestamp > intervalTs.Unix() {
		sendAlert(title, text, url)
	}
}

// Send the alert out
func sendAlert(title string, text string, url string) {
	// Setup
	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", smtpTo)
	m.SetHeader("Subject", "Reddit Notify: Found match ("+title+")")
	m.SetBody("text/plain", "Keyword: "+keyword+"\n\n\n"+title+"\n\n\n"+text)

	// Send via gomail
	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
	}
}
