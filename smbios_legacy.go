// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bytes"
	"errors"
	"io"
)

// Errors for entry points.
var (
	ErrAnchorNotFound = errors.New("SMBIOS anchor _SM_ or _SM3_ not found")
)

// getMemBase searches _SM_ or _SM3_ tag in the given memory range.
func getMemBase(r io.ReaderAt, start, end int64) (addr int64, size int64, err error) {
	b := make([]byte, 5)
	for base := start; base < end-5; base++ {
		if _, err := io.ReadFull(io.NewSectionReader(r, base, 5), b); err != nil {
			return 0, 0, err
		}
		if bytes.Equal(b[:4], anchor32) {
			return base, smbios2HeaderSize, nil
		}
		if bytes.Equal(b, anchor64) {
			return base, smbios3HeaderSize, nil
		}
	}
	return 0, 0, ErrAnchorNotFound
}
