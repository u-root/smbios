// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bytes"
	"errors"
	"io"
	"runtime"
	"testing"

	"github.com/hugelgupf/vmtest/guest"
)

func TestLegacy(t *testing.T) {
	for _, tt := range []struct {
		// in
		b     []byte
		start int64
		end   int64
		// out
		addr int64
		size int64
		err  error
	}{
		{
			b:     bytes.Repeat([]byte{0}, 100),
			start: 10,
			end:   100,
			err:   ErrAnchorNotFound,
		},
		{
			b:     []byte{0, '_', 'M', 'S', '_', 0, 0, '_', 'S', 'M', '_', 0, 0, 0, 0, 0},
			start: 1,
			end:   15,
			addr:  7,
			size:  smbios2HeaderSize,
		},
		{
			b:     []byte{0, '_', 'M', 'S', '_', 0, 0, '_', 'S', 'M', '3', '_', 0, 0, 0, 0, 0},
			start: 1,
			end:   15,
			addr:  7,
			size:  smbios3HeaderSize,
		},
		{
			b:     []byte{0, '_', 'M', 'S', '_', 0, 0, '_', 'S'},
			start: 1,
			end:   15,
			err:   io.ErrUnexpectedEOF,
		},
	} {
		t.Run("", func(t *testing.T) {
			addr, size, err := getMemBase(bytes.NewReader(tt.b), tt.start, tt.end)
			if !errors.Is(err, tt.err) {
				t.Errorf("getMemBase = %v, want %v", err, tt.err)
			}
			if addr != tt.addr {
				t.Errorf("getMemBase = addr %x, want %x", addr, tt.addr)
			}
			if size != tt.size {
				t.Errorf("getMemBase = size %x, want %x", size, tt.size)
			}
		})
	}
}

func TestSMBIOSLegacyQEMU(t *testing.T) {
	guest.SkipIfNotInVM(t)
	if runtime.GOARCH != "amd64" {
		t.Skipf("test not supported on %s", runtime.GOARCH)
	}

	base, size, err := EntryBaseFromLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if base == 0 {
		t.Errorf("SMBIOSLegacy() does not get SMBIOS base")
	}

	if size != smbios2HeaderSize && size != smbios3HeaderSize {
		t.Errorf("BaseLegacy(): %v, want '%v' or '%v'", size, smbios2HeaderSize, smbios3HeaderSize)
	}
}
