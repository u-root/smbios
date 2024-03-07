// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"testing"
)

func TestSMBIOSEFISMBIOS2(t *testing.T) {
	systabPath = "testdata/smbios2_systab"
	base, size, err := BaseEFI()
	if err != nil {
		t.Fatal(err)
	}

	var want int64 = 0x12345678

	if base != want {
		t.Errorf("BaseEFI(): 0x%x, want 0x%x", base, want)
	}
	if size != smbios2HeaderSize {
		t.Errorf("BaseEFI(): 0x%x, want 0x%x ", size, smbios2HeaderSize)
	}
}

func TestSMBIOSEFISMBIOS3(t *testing.T) {
	systabPath = "testdata/smbios3_systab"
	base, size, err := BaseEFI()
	if err != nil {
		t.Fatal(err)
	}

	var want int64 = 0x12345678

	if base != want {
		t.Errorf("BaseEFI(): 0x%x, want 0x%x", base, want)
	}
	if size != smbios3HeaderSize {
		t.Errorf("BaseEFI(): 0x%x, want 0x%x ", size, smbios3HeaderSize)
	}
}

func TestSMBIOSEFINotFound(t *testing.T) {
	systabPath = "testdata/systab_NOT_FOUND"
	_, _, err := BaseEFI()
	if err == nil {
		t.Errorf("BaseEFI(): nil , want error")
	}
}

func TestSMBIOSEFIInvalid(t *testing.T) {
	systabPath = "testdata/invalid_systab"
	_, _, err := BaseEFI()
	if err == nil {
		t.Errorf("BaseEFI(): nil , want error")
	}
}
