package main

import (
	"embed"
	"html/template"
	"log"

	"github.com/riraum/si-cheong/db"
	"github.com/riraum/si-cheong/http"
	"github.com/riraum/si-cheong/security"
)

//go:embed static/*
var static embed.FS
var t = template.Must(template.ParseFS(static, "static/*"))

func main() {
	log.Print("Hello si-cheong user")

	key, err := security.NewEncryptionKey()
	if err != nil {
		log.Fatalf("key fail: %v", err)
	}

	d, err := db.New("./sq.db")
	if err != nil {
		log.Fatalf("Failed to create new db %v", err)
	}

	if err = d.Fill(); err != nil {
		log.Fatalf("error filling posts into db: %v", err)
	}

	s := http.Server{
		EmbedRootDir: static,
		DB:           d,
		Template:     t,
		Key:          key,
	}

	mux := s.SetupMux()
	http.Run(mux)
}
