// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/u-root/smbios"
)

func TestBIOSCharsString(t *testing.T) {
	tests := []struct {
		name string
		val  uint64
		want string
	}{
		{
			name: "Reserved",
			val:  0x1,
			want: "\t\tReserved",
		},
		{
			name: "Every Option",
			val:  0xFFFFFFFFFFFF,
			want: `		Reserved
		Reserved
		Unknown
		BIOS characteristics not supported
		ISA is supported
		MCA is supported
		EISA is supported
		PCI is supported
		PC Card (PCMCIA) is supported
		PNP is supported
		APM is supported
		BIOS is upgradeable
		BIOS shadowing is allowed
		VLB is supported
		ESCD support is available
		Boot from CD is supported
		Selectable boot is supported
		BIOS ROM is socketed
		Boot from PC Card (PCMCIA) is supported
		EDD is supported
		Japanese floppy for NEC 9800 1.2 MB is supported (int 13h)
		Japanese floppy for Toshiba 1.2 MB is supported (int 13h)
		5.25"/360 kB floppy services are supported (int 13h)
		5.25"/1.2 MB floppy services are supported (int 13h)
		3.5"/720 kB floppy services are supported (int 13h)
		3.5"/2.88 MB floppy services are supported (int 13h)
		Print screen service is supported (int 5h)
		8042 keyboard services are supported (int 9h)
		Serial services are supported (int 14h)
		Printer services are supported (int 17h)
		CGA/mono video services are supported (int 10h)
		NEC PC-98`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BIOSChars(tt.val).String()
			if got != tt.want {
				t.Errorf("BIOSChars().String(): '%s', want '%s'", got, tt.want)
			}
		})
	}
}

func TestBIOSCharsExt1String(t *testing.T) {
	tests := []struct {
		name string
		val  uint8
		want string
	}{
		{
			name: "All options Ext1",
			val:  0xFF,
			want: `		ACPI is supported
		USB legacy is supported
		AGP is supported
		I2O boot is supported
		LS-120 boot is supported
		ATAPI Zip drive boot is supported
		IEEE 1394 boot is supported
		Smart battery is supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testVal = BIOSCharsExt1(tt.val)

			resultString := testVal.String()

			if resultString != tt.want {
				t.Errorf("BIOSCharsExt1().String(): '%s', want '%s'", resultString, tt.want)
			}
		})
	}
}

func TestBIOSCharsExt2String(t *testing.T) {
	tests := []struct {
		name string
		val  uint8
		want string
	}{
		{
			name: "All options Ex2",
			val:  0xFF,
			want: `		BIOS boot specification is supported
		Function key-initiated network boot is supported
		Targeted content distribution is supported
		UEFI is supported
		System is a virtual machine`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testVal = BIOSCharsExt2(tt.val)

			resultString := testVal.String()

			if resultString != tt.want {
				t.Errorf("BIOSCharsExt2().String(): '%s', want '%s'", resultString, tt.want)
			}
		})
	}
}

func TestBIOSInfoString(t *testing.T) {
	tests := []struct {
		name string
		val  BIOSInfo
		want string
	}{
		{
			name: "Valid BIOSInfo",
			val: BIOSInfo{
				Vendor:                 "u-root",
				Version:                "1.0",
				StartingAddressSegment: 0x4,
				ReleaseDate:            "2021/11/23",
				ROMSize:                8,
				Characteristics:        BIOSChars(0x8),
				CharacteristicsExt1:    BIOSCharsExt1(0x4),
				CharacteristicsExt2:    BIOSCharsExt2(0x2),
				BIOSMajor:              0xff,
				BIOSMinor:              0xff,
				ECMajor:                0xff,
				ECMinor:                0xff,
				ExtendedROMSize:        0,
			},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Vendor: u-root
	Version: 1.0
	Release Date: 2021/11/23
	Address: 0x00040
	Runtime Size: 1048512 bytes
	ROM Size: 576 kB
	Characteristics:
		BIOS characteristics not supported
		AGP is supported
		Function key-initiated network boot is supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.String() != tt.want {
				t.Errorf("String(): %s, want %s", tt.val.String(), tt.want)
			}
		})
	}
}

func TestROMSizeBytes(t *testing.T) {
	for _, tt := range []struct {
		name string
		val  BIOSInfo
		want uint64
	}{
		{
			name: "Rom Size 0xFF",
			val: BIOSInfo{
				ROMSize: 0xFF,
			},
			want: 0x1000000,
		},
		{
			name: "Rom Size 0xAB",
			val: BIOSInfo{
				ROMSize: 0xAB,
			},
			want: 0xAC0000,
		},
		{
			name: "Big Ext Size",
			val: BIOSInfo{
				ROMSize:         0xFF,
				ExtendedROMSize: 0xFFFF,
			},
			want: 0x3FFF,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			romSize := tt.val.ROMSizeBytes()
			if romSize != tt.want {
				t.Errorf("GetROMSizeBytes = %v, want %v", romSize, tt.want)
			}
		})
	}
}

func TestParseBIOSInfo(t *testing.T) {
	for _, tt := range []struct {
		name  string
		table *smbios.Table
		want  *BIOSInfo
		err   error
	}{
		{
			name: "BIOS Info",
			table: &smbios.Table{
				Header: smbios.Header{
					Length: 22,
					Type:   smbios.TableTypeBIOSInfo,
				},
				Data: []byte{
					0x00, // vendor string
					0x01, // version string
					0x02, 0x03,
					0x02,                                           // release date string
					0x00,                                           // rom size
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // bios chars
					0x00,
					0x00,
					0x0b, 0x0c, // BIOS major, minor
					0x0d, 0x0e, // EC major minor
					0x00, 0x10, // rom size
				},
				Strings: []string{
					"version!",
					"release date!",
				},
			},
			want: &BIOSInfo{
				Header: smbios.Header{
					Length: 22,
					Type:   smbios.TableTypeBIOSInfo,
				},
				Version:                "version!",
				StartingAddressSegment: 0x302,
				ReleaseDate:            "release date!",
				Characteristics:        0x1,
				BIOSMajor:              0xb,
				BIOSMinor:              0xc,
				ECMajor:                0xd,
				ECMinor:                0xe,
				ExtendedROMSize:        0x1000,
			},
		},
		{
			name: "short",
			table: &smbios.Table{
				Header: smbios.Header{
					Length: 4,
					Type:   smbios.TableTypeBIOSInfo,
				},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "invalid type",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeCacheInfo,
				},
			},
			err: ErrUnexpectedTableType,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBIOSInfo(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseBIOSInfo() = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBIOSInfo = %v, want %v", got, tt.want)
			}
		})
	}
}
