// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/arnesssr/Spyber/internal/app"
	"github.com/arnesssr/Spyber/internal/infra/commoncrawl"
	"github.com/arnesssr/Spyber/internal/infra/countryfinders"
	"github.com/arnesssr/Spyber/internal/infra/htmlparse"
	"github.com/arnesssr/Spyber/internal/infra/httpfetch"
	"github.com/arnesssr/Spyber/internal/infra/overpass"
	"github.com/arnesssr/Spyber/internal/infra/storeconfig"
	"github.com/arnesssr/Spyber/internal/infra/websearch"
	"github.com/arnesssr/Spyber/internal/interface/web"
	"github.com/arnesssr/Spyber/internal/version"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8091", "address to listen on")
	storePath := flag.String("store", "", "local JSON store path")
	flag.Parse()
	if *storePath != "" {
		_ = os.Setenv("SPYBER_STORE", *storePath)
	}

	store, err := storeconfig.Open()
	if err != nil {
		log.Fatal(err)
	}
	finder := countryfinders.New(
		websearch.New(os.Getenv("SPYBER_WEBSEARCH_ENDPOINT")),
		overpass.New(os.Getenv("SPYBER_OVERPASS_ENDPOINT")),
		commoncrawl.New(os.Getenv("SPYBER_COMMONCRAWL_INDEX")),
	)
	application := app.New(store.Store, httpfetch.New(), htmlparse.New()).WithCountryFinder(finder)
	if err := application.Init(context.Background()); err != nil {
		log.Fatal(err)
	}

	server := web.New(application, web.Config{
		AdminToken: os.Getenv("SPYBER_ADMIN_TOKEN"),
	})
	log.Printf("spyberd %s listening on http://%s using %s", version.Version, *addr, store.Label)
	log.Fatal(http.ListenAndServe(*addr, server))
}
