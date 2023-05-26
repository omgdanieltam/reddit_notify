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
	subreddit, smtpServer, smtpTo, smtpFrom, smtpUsername, smtpPassword string
	keywords                                                            []string
	interval, smtpPort                                                  int
)

func main() {
	// Load config
	cfg, err := ini.ShadowLoad("config.ini")

	if err != nil {
		quitConfigParseError(err.Error())
	}

	// Parse and check config values
	if cfg.Section("app").HasKey("subreddit") {
		subreddit = cfg.Section("app").Key("subreddit").String()
		printConfig("subreddit", subreddit)
	} else {
		quitConfigParseError("Missing 'subreddit'")
	}

	if cfg.Section("app").HasKey("interval") {
		// default to 5 minutes
		interval = cfg.Section("app").Key("interval").MustInt(5)
		printConfig("interval", strconv.Itoa(interval))
	} else {
		quitConfigParseError("Missing 'interval'")
	}

	if cfg.Section("app").HasKey("keyword") {
		keywords = cfg.Section("app").Key("keyword").ValueWithShadows()
		for _, keys := range keywords {
			printConfig("keyword", keys)
		}
	} else {
		quitConfigParseError("Missing 'keyword'")
	}

	if cfg.Section("smtp").HasKey("smtp_server") {
		smtpServer = cfg.Section("smtp").Key("smtp_server").String()
		printConfig("smtp_server", smtpServer)
	} else {
		quitConfigParseError("Missing 'smtp_server'")
	}

	if cfg.Section("smtp").HasKey("smtp_port") {
		// default to port 25
		smtpPort = cfg.Section("smtp").Key("smtp_port").MustInt(25)
		printConfig("smtp_port", strconv.Itoa(smtpPort))
	} else {
		quitConfigParseError("Missing 'smtp_port'")
	}

	if cfg.Section("smtp").HasKey("smtp_username") {
		smtpUsername = cfg.Section("smtp").Key("smtp_username").String()
		printConfig("smtp_username", smtpUsername)
	} else {
		quitConfigParseError("Missing 'smtp_username'")
	}

	if cfg.Section("smtp").HasKey("smtp_password") {
		smtpPassword = cfg.Section("smtp").Key("smtp_password").String()
		printConfig("smtp_password", "<redacted>")
	} else {
		quitConfigParseError("Missing 'smtp_password'")
	}

	if cfg.Section("smtp").HasKey("smtp_to") {
		smtpTo = cfg.Section("smtp").Key("smtp_to").String()
		printConfig("smtp_to", smtpTo)
	} else {
		quitConfigParseError("Missing 'smtp_to'")
	}

	if cfg.Section("smtp").HasKey("smtp_from") {
		smtpFrom = cfg.Section("smtp").Key("smtp_from").String()
		printConfig("smtp_from", smtpFrom)
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

			for _, keys := range keywords {
				// Check for keywords
				alert := false
				if compareToKeywords(strings.ToLower(title), strings.ToLower(keys)) {
					alert = true
				} else if compareToKeywords(strings.ToLower(text), strings.ToLower(keys)) {
					alert = true
				}

				// Send alert if keyword found
				if alert {
					url, _ := jsonparser.GetString(body, "data", "children", "["+index+"]", "data", "url")
					timestamp, _ := jsonparser.GetFloat(body, "data", "children", "["+index+"]", "data", "created_utc")
					validateAlert(title, text, url, int64(timestamp), keys)
				}
			}
		}

		// Sleep for interval time
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func printConfig(key string, value string) {
	fmt.Println("Loaded "+key+": ", value)
}

func quitConfigParseError(msg string) {
	fmt.Println("Error parsing config.ini: ", msg)
	os.Exit(1)
}

func compareToKeywords(text string, keyword string) bool {
	// Split keywords on commas
	keys := strings.Split(keyword, ",")
	found := false

	// Check to ensure it contains ALL the keywords
	for _, key := range keys {
		if strings.Contains(text, strings.TrimSpace(key)) {
			found = true
		} else {
			found = false
		}
	}

	return found
}

// Validate the alert to ensure that it needs to be sent
func validateAlert(title string, text string, url string, timestamp int64, keyword string) {
	// Get timestamp of interval period
	currentTs := time.Now()
	intervalTs := currentTs.Add(-time.Minute * time.Duration(interval))

	// Only send alert if it's newer than interval time period
	if timestamp > intervalTs.Unix() {
		sendAlert(title, text, url, keyword)
	}
}

// Send the alert out
func sendAlert(title string, text string, url string, keyword string) {
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
