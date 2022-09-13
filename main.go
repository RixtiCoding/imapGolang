package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func main() {
	var OTPcode string
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("[imap.gmail.com]:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("-", "-"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 15
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	section := imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem(), imap.FetchRFC822Header}

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var vevemail *imap.Message
	for msg := range messages {
		addresses := msg.Envelope.From
		for _, address := range addresses {
			name := address.Address()
			if name == "no-reply@veve.me" && strings.Contains(msg.Envelope.Subject, "2FA") {
				vevemail = msg
				break
			}
		}
	}
	if vevemail != nil {
		fmt.Printf("OTP for %v found\n", vevemail.Envelope.To[0].Address())
		r := vevemail.GetBody(&section)

		var mailtext []byte
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			switch x := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				mailtext = b
				_ = x
			}
		}

		re := regexp.MustCompile(`\>(.*?)\</span`).FindAllStringSubmatch(string(mailtext), -1)
		OTPcode = re[1][1]
		log.Printf("Fetched the OTPCODE! => %v", OTPcode)
	} else {
		log.Println("No OTP found! please fetch the code manually")
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
