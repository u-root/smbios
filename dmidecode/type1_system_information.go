// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/u-root/smbios"
)

// SystemInfo is defined in DSP0134 7.2.
type SystemInfo struct {
	smbios.Header `smbios:"-"`
	Manufacturer  string     // 04h
	ProductName   string     // 05h
	Version       string     // 06h
	SerialNumber  string     // 07h
	UUID          UUID       // 08h
	WakeupType    WakeupType // 18h
	SKUNumber     string     // 19h
	Family        string     // 1Ah
}

// ParseSystemInfo parses a generic Table into SystemInfo.
func ParseSystemInfo(t *smbios.Table) (*SystemInfo, error) {
	if t.Type != smbios.TableTypeSystemInfo {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	if t.Len() < 8 {
		return nil, fmt.Errorf("%w: system info table must be at least %d bytes", io.ErrUnexpectedEOF, 8)
	}
	si := &SystemInfo{Header: t.Header}
	if _, err := parseStruct(t, 0 /* off */, false /* complete */, si); err != nil {
		return nil, err
	}
	return si, nil
}

func (si *SystemInfo) String() string {
	lines := []string{
		si.Header.String(),
		fmt.Sprintf("Manufacturer: %s", smbiosStr(si.Manufacturer)),
		fmt.Sprintf("Product Name: %s", smbiosStr(si.ProductName)),
		fmt.Sprintf("Version: %s", smbiosStr(si.Version)),
		fmt.Sprintf("Serial Number: %s", smbiosStr(si.SerialNumber)),
		fmt.Sprintf("UUID: %s", si.UUID),
		fmt.Sprintf("Wake-up Type: %s", si.WakeupType),
		fmt.Sprintf("SKU Number: %s", smbiosStr(si.SKUNumber)),
		fmt.Sprintf("Family: %s", smbiosStr(si.Family)),
	}
	return strings.Join(lines, "\n\t")
}

// UUID is defined in DSP0134 7.2.1.
type UUID [16]byte

func (u UUID) String() string {
	if bytes.Equal(u[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		return "Not Settable"
	}
	if bytes.Equal(u[:], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
		return "Not Present"
	}
	// Note: First three fields use LE byte order, last two use BE (network).
	// Reasons for this are described in 7.2.1 (basically: historic).
	// dmidecode(8) only does this for 2.6+ SMBIOS versions but we don't make that distinction.
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		u[3], u[2], u[1], u[0],
		u[5], u[4],
		u[7], u[6],
		u[8], u[9],
		u[10], u[11], u[12], u[13], u[14], u[15],
	)
}

// WakeupType is defined in DSP0134 7.2.2.
type WakeupType uint8

// WakeupType values are defined in DSP0134 7.2.2.
const (
	WakeupTypeReserved        WakeupType = 0x00 // Reserved
	WakeupTypeOther           WakeupType = 0x01 // Other
	WakeupTypeUnknown         WakeupType = 0x02 // Unknown
	WakeupTypeAPMTimer        WakeupType = 0x03 // APM Timer
	WakeupTypeModemRing       WakeupType = 0x04 // Modem Ring
	WakeupTypeLANRemote       WakeupType = 0x05 // LAN Remote
	WakeupTypePowerSwitch     WakeupType = 0x06 // Power Switch
	WakeupTypePCIPME          WakeupType = 0x07 // PCI PME#
	WakeupTypeACPowerRestored WakeupType = 0x08 // AC Power Restored
)

var wakeupStrings = map[WakeupType]string{
	WakeupTypeReserved:        "Reserved",
	WakeupTypeOther:           "Other",
	WakeupTypeUnknown:         "Unknown",
	WakeupTypeAPMTimer:        "APM Timer",
	WakeupTypeModemRing:       "Modem Ring",
	WakeupTypeLANRemote:       "LAN Remote",
	WakeupTypePowerSwitch:     "Power Switch",
	WakeupTypePCIPME:          "PCI PME#",
	WakeupTypeACPowerRestored: "AC Power Restored",
}

func (v WakeupType) String() string {
	if name, ok := wakeupStrings[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}
