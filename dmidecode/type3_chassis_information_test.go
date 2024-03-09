// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/u-root/smbios"
)

func TestChassisInfoString(t *testing.T) {
	tests := []struct {
		name string
		val  ChassisInfo
		want string
	}{
		{
			name: "Full Information",
			val: ChassisInfo{
				Header: smbios.Header{
					Type:   smbios.TableTypeChassisInfo,
					Length: 0xe,
				},
				Manufacturer:       "The Ancients",
				Type:               ChassisTypeAllInOne,
				Version:            "One",
				SerialNumber:       "TheAncients-01",
				AssetTagNumber:     "Two",
				BootupState:        ChassisStateSafe,
				PowerSupplyState:   ChassisStateSafe,
				ThermalState:       ChassisStateNonrecoverable,
				SecurityStatus:     ChassisSecurityStatusUnknown,
				OEMInfo:            0xABCD0123,
				Height:             3,
				NumberOfPowerCords: 1,
				ContainedElements: []ChassisContainedElement{
					{
						Type: ChassisElementType(8),
						Min:  3,
						Max:  11,
					},
					{
						Type: ChassisElementType(0),
						Min:  0,
						Max:  1,
					},
				},
				SKUNumber: "Four",
			},
			want: `Handle 0x0000, DMI type 3, 14 bytes
Chassis Information
	Manufacturer: The Ancients
	Type: All In One
	Lock: Not Present
	Version: One
	Serial Number: TheAncients-01
	Asset Tag: Two
	Boot-up State: Safe
	Power Supply State: Safe
	Thermal State: Non-recoverable
	Security Status: Unknown
	OEM Information: 0xABCD0123
	Height: 3 U
	Number Of Power Cords: 1
	Contained Elements: 2
		Memory Module 3-11
		0x0 0-1`,
		}, {
			name: "Minimal Information",
			val: ChassisInfo{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Manufacturer:   "The Ancients",
				Type:           ChassisTypeTower,
				Version:        "Two",
				SerialNumber:   "TheAncients-02",
				AssetTagNumber: "Three",
			},
			want: `Handle 0x0000, DMI type 3, 0 bytes
Chassis Information
	Manufacturer: The Ancients
	Type: Tower
	Lock: Not Present
	Version: Two
	Serial Number: TheAncients-02
	Asset Tag: Three`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.val.String()
			if result != tt.want {
				t.Errorf("ChassisInfo().String(): %v, want '%v'", result, tt.want)
			}
		})
	}
}

func TestChassisTypeString(t *testing.T) {
	testResults := []string{
		"Other",
		"Unknown",
		"Desktop",
		"Low Profile Desktop",
		"Pizza Box",
		"Mini Tower",
		"Tower",
		"Portable",
		"Laptop",
		"Notebook",
		"Hand Held",
		"Docking Station",
		"All In One",
		"Sub Notebook",
		"Space-saving",
		"Lunch Box",
		"Main Server Chassis",
		"Expansion Chassis",
		"Sub Chassis",
		"Bus Expansion Chassis",
		"Peripheral Chassis",
		"RAID Chassis",
		"Rack Mount Chassis",
		"Sealed-case PC",
		"Multi-system",
		"CompactPCI",
		"AdvancedTCA",
		"Blade",
		"Blade Chassis",
		"Tablet",
		"Convertible",
		"Detachable",
		"IoT Gateway",
		"Embedded PC",
		"Mini PC",
		"Stick PC",
	}

	for id := range testResults {
		val := ChassisType(id + 1)
		if val.String() != testResults[id] {
			t.Errorf("ChassisType().String() = %s want %s", val.String(), testResults[id])
		}
	}
}

func TestParseChassisInfo(t *testing.T) {
	for _, tt := range []struct {
		name  string
		table *smbios.Table
		want  *ChassisInfo
		err   error
	}{
		{
			name: "Invalid Type",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeBIOSInfo,
				},
			},
			err: ErrUnexpectedTableType,
		},
		{
			name: "Required fields are missing",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Data: []byte{},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "Error parsing structure",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Data: []byte{
					0x00, 0x01, 0x02, 0x03, 0x04,
					0x05, 0x06, 0x07,
					0x08,
					0x09, 0x0a,
					0x0b,
					0x0c, // element size
					0x0d, 0x0e, 0x0f, 0x10,
					0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
					0x1a,
				},
			},
			err: os.ErrInvalid,
		},
		{
			name: "Parse valid SystemInfo",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Data: []byte{
					0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, // chassis states
					0x00,                   // security status
					0x00, 0x00, 0x00, 0x00, // oem info
					0x00, // height
					0x00, // num powercords
					0x03, // num elements
					0x03, // element size
					0x11, 0x12, 0x13,
					0x14, 0x15, 0x16,
					0x17, 0x18, 0x19,
					0x01, // SKU
				},
				Strings: []string{
					"SKU!",
				},
			},
			want: &ChassisInfo{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				ContainedElements: []ChassisContainedElement{
					{Type: 0x11, Min: 0x12, Max: 0x13},
					{Type: 0x14, Min: 0x15, Max: 0x16},
					{Type: 0x17, Min: 0x18, Max: 0x19},
				},
				SKUNumber: "SKU!",
			},
		},
		{
			name: "Parse valid SystemInfo",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Data: []byte{
					0x00,
					0x01, // type
					0x00,
					0x00,
					0x00,
					0x03, 0x03, 0x03, // states
					0x03,                   // security
					0x34, 0x12, 0x00, 0x00, // oem info
					0x00, // height
					0x00, // num power
					0x00, // num elems
					0x00, // elem size
				},
			},
			want: &ChassisInfo{
				Header: smbios.Header{
					Type: smbios.TableTypeChassisInfo,
				},
				Type:             0x01,
				BootupState:      0x03,
				PowerSupplyState: 0x03,
				ThermalState:     0x03,
				SecurityStatus:   0x03,
				OEMInfo:          0x1234,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChassisInfo(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseChassisInfo = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseChassisInfo = %v, want %v", got, tt.want)
			}
		})
	}
}
