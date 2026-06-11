// SPDX-License-Identifier: AGPL-3.0-only

package storeconfig

import (
	"os"

	"github.com/waymore/spyber/internal/infra/localstore"
	"github.com/waymore/spyber/internal/infra/pgstore"
	"github.com/waymore/spyber/internal/ports"
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
	return StoreConfig{Store: localstore.New(Path()), Label: Path()}, nil
}

func Path() string {
	if value := os.Getenv("SPYBER_STORE"); value != "" {
		return value
	}
	return ".spyber/spyber.json"
}
