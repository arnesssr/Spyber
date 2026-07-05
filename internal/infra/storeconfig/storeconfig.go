// SPDX-License-Identifier: AGPL-3.0-only

package storeconfig

import (
	"fmt"
	"os"

	"github.com/arnesssr/Spyber/internal/infra/localstore"
	"github.com/arnesssr/Spyber/internal/infra/pgstore"
	"github.com/arnesssr/Spyber/internal/ports"
)

type StoreConfig struct {
	Store ports.Store
	Label string
}

func Open() (StoreConfig, error) {
	if databaseURL := os.Getenv("SPYBER_DATABASE_URL"); databaseURL != "" {
		store, err := pgstore.New(databaseURL)
		if err != nil {
			return StoreConfig{}, err
		}
		return StoreConfig{Store: store, Label: "postgresql"}, nil
	}
	if path := Path(); path != "" {
		return StoreConfig{Store: localstore.New(path), Label: "dev-json:" + path}, nil
	}
	return StoreConfig{}, fmt.Errorf("SPYBER_DATABASE_URL is required; set SPYBER_STORE only for development JSON runs")
}

func Path() string {
	if value := os.Getenv("SPYBER_STORE"); value != "" {
		return value
	}
	return ""
}
