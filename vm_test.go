// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race
// +build !race

package smbios

import (
	"runtime"
	"testing"
	"time"

	"github.com/hugelgupf/vmtest/govmtest"
	"github.com/hugelgupf/vmtest/guest"
	"github.com/hugelgupf/vmtest/qemu"
)

func TestIntegration(t *testing.T) {
	qemu.SkipIfNotArch(t, qemu.ArchAMD64)

	govmtest.Run(t, "vm",
		govmtest.WithPackageToTest("github.com/u-root/smbios"),
		govmtest.WithQEMUFn(
			qemu.WithVMTimeout(time.Minute),
			qemu.ArbitraryArgs("-smbios", "type=2,manufacturer=u-root"),
		),
	)
}

func TestLegacyQEMU(t *testing.T) {
	guest.SkipIfNotInVM(t)
	if runtime.GOARCH != "amd64" {
		t.Skip("amd64 only")
	}

	addr, size, err := EntryBaseFromLegacy()
	if err != nil {
		t.Error(err)
	}
	if size != 0x18 {
		t.Errorf("QEMU expected to use SMBIOS3")
	}
	t.Logf("addr: %#x", addr)
	t.Logf("size: %#x", size)

	gaddr, gsize, err := EntryBase()
	if err != nil {
		t.Error(err)
	}
	if gsize != 0x18 {
		t.Errorf("QEMU expected to use SMBIOS3")
	}
	if gaddr != addr {
		t.Errorf("EntryBase address (%#x) not same as EntryBaseFromLegacy (%#x)", gaddr, addr)
	}
}
