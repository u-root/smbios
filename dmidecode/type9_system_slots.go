// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"fmt"
	"io"

	"github.com/u-root/smbios"
)

// SystemSlots is defined in DSP0134 7.10.
type SystemSlots struct {
	smbios.Table
	SlotDesignation      string // 04h
	SlotType             uint8  // 05h
	SlotDataBusWidth     uint8  // 06h
	CurrentUsage         uint8  // 07h
	SlotLength           uint8  // 08h
	SlotID               uint16 // 09h
	SlotCharacteristics1 uint8  // 0Bh
	SlotCharacteristics2 uint8  // 0Ch
	SegmentGroupNumber   uint16 // 0Dh
	BusNumber            uint8  // 0Fh
	DeviceFunctionNumber uint8  // 10h
	DataBusWidth         uint8  // 11h
	// TODO: Peer grouping count and Peer groups are omitted for now
}

// ParseSystemSlots parses a generic smbios.Table into SystemSlots.
func ParseSystemSlots(t *smbios.Table) (*SystemSlots, error) {
	if t.Type != smbios.TableTypeSystemSlots {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	if t.Len() < 0xb {
		return nil, fmt.Errorf("%w: system slots table must be at least %d bytes", io.ErrUnexpectedEOF, 0xb)
	}

	s := &SystemSlots{Table: *t}
	_, err := parseStruct(t, 0 /* off */, false /* complete */, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Typ implements Table.Typ.
func (s SystemSlots) Typ() smbios.TableType {
	return smbios.TableTypeSystemSlots
}
