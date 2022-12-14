package main

import (
	"log"

	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
)

func serv() {
	// Create a memory backend
	be := memory.New()

	// Create a new server
	s := server.New(be)
	s.Addr = ":1143"
	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Println("Starting IMAP server at localhost:1143")
	go s.ListenAndServe()

}
