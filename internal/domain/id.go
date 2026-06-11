// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type ID string

func NewID(prefix string) ID {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ID(prefix + "_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", ""))
	}
	return ID(prefix + "_" + hex.EncodeToString(b[:]))
}

func (id ID) String() string {
	return string(id)
}
