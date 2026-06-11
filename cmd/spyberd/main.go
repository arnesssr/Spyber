// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/infra/commoncrawl"
	"github.com/waymore/spyber/internal/infra/countryfinders"
	"github.com/waymore/spyber/internal/infra/htmlparse"
	"github.com/waymore/spyber/internal/infra/httpfetch"
	"github.com/waymore/spyber/internal/infra/localstore"
	"github.com/waymore/spyber/internal/infra/overpass"
	"github.com/waymore/spyber/internal/interface/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on")
	storePath := flag.String("store", defaultStore(), "local store path")
	flag.Parse()

	store := localstore.New(*storePath)
	finder := countryfinders.New(
		overpass.New(os.Getenv("SPYBER_OVERPASS_ENDPOINT")),
		commoncrawl.New(os.Getenv("SPYBER_COMMONCRAWL_INDEX")),
	)
	application := app.New(store, httpfetch.New(), htmlparse.New()).WithCountryFinder(finder)
	if err := application.Init(context.Background()); err != nil {
		log.Fatal(err)
	}

	server := web.New(application, web.Config{
		AdminToken: os.Getenv("SPYBER_ADMIN_TOKEN"),
	})
	log.Printf("spyberd listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, server))
}

func defaultStore() string {
	if value := os.Getenv("SPYBER_STORE"); value != "" {
		return value
	}
	return ".spyber/spyber.json"
}
