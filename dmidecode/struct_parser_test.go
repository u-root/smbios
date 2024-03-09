// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/u-root/smbios"
)

func checkError(got error, want error) bool {
	if got != nil && want != nil {
		if got.Error() == want.Error() {
			return true
		}
	}

	return errors.Is(got, want)
}

type UnknownTypes struct {
	smbios.Table
	SupportedField   uint64
	UnsupportedField float32
}

func TestParseStructUnsupported(t *testing.T) {
	buffer := []byte{
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x00, 0x01, 0x02, 0x03,
	}

	want := "unsupported type float32"
	table := smbios.Table{
		Data: buffer,
	}
	UnknownType := &UnknownTypes{
		Table: table,
	}

	off, err := parseStruct(&table, 0, false, UnknownType)
	if err == nil {
		t.Errorf("TestParseStructUnsupported : parseStruct() = %d, '%v' want: %q", off, err, want)
	} else if !strings.Contains(err.Error(), want) {
		t.Errorf("TestParseStructUnsupported : parseStruct() = %d, '%v' want: %q", off, err, want)
	}
}

func TestParseStruct(t *testing.T) {
	type foobar struct {
		Foo uint8 `smbios:"default=0xe"`
	}
	type someStruct struct {
		Off0  uint64
		Off8  uint8
		Off9  string
		_     uint8 `smbios:"-"`
		Off10 uint16
		Off14 uint8 `smbios:"skip=2"`
		_     uint8 `smbios:"-"`
		Off15 uint8 `smbios:"skip=2,default=0x1"`
		Off17 uint8 `smbios:"default=0xf"`
		Off18 foobar
	}

	for _, tt := range []struct {
		table *smbios.Table
		value any
		err   error
		want  any
	}{
		{
			table: &smbios.Table{
				Data: []byte{
					0x1, 0x0, 0x0, 0x0,
					0x0, 0x0, 0x0, 0x0,
					0xff,     // Off8
					0x1,      // Off9
					0x2, 0x1, // Off10
					0xff, 0xff, // skipped
					0x5, // Off14
				},
				Strings: []string{
					"foobar",
				},
			},
			value: &someStruct{},
			want: &someStruct{
				Off0:  0x01,
				Off8:  0xff,
				Off9:  "foobar",
				Off10: 0x102,
				Off14: 0x05,
				Off15: 0x1,
				Off17: 0x0f,
				Off18: foobar{
					Foo: 0xe,
				},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			if _, err := parseStruct(tt.table, 0, false, tt.value); !errors.Is(err, tt.err) {
				t.Errorf("parseStruct = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(tt.value, tt.want) {
				t.Errorf("parseStruct = %v, want %v", tt.value, tt.want)
			}
		})
	}
}

func TestParseStructWithTPMDevice(t *testing.T) {
	tests := []struct {
		name     string
		buffer   []byte
		strings  []string
		complete bool
		want     TPMDevice
		wantErr  error
	}{
		{
			name: "Type43TPMDevice",
			buffer: []byte{
				0x00, 0x00, 0x00, 0x00, // VendorID
				0x02,       // Major
				0x03,       // Minor
				0x01, 0x00, // FirmwareVersion1
				0x02, 0x00, // FirmwareVersion1
				0x00, 0x00, 0x00, 0x00, // FirmwareVersion2
				0x01,                   // String Index
				1 << 3,                 // Characteristics
				0x78, 0x56, 0x34, 0x12, // OEMDefined
			},
			strings:  []string{"Test TPM"},
			complete: false,
			want: TPMDevice{
				VendorID:         [4]byte{0x00, 0x00, 0x00, 0x00},
				MajorSpecVersion: 2,
				MinorSpecVersion: 3,
				FirmwareVersion1: 0x00020001,
				FirmwareVersion2: 0x00000000,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristicsFamilyConfigurableViaFirmwareUpdate,
				OEMDefined:       0x12345678,
			},
			wantErr: nil,
		},
		{
			name: "Type43TPMDevice Incomplete",
			buffer: []byte{
				0x00, 0x00, 0x00, 0x00, // VendorID
				0x02,       // Major
				0x03,       // Minor
				0x01, 0x00, // FirmwareVersion1
				0x02, 0x00, // FirmwareVersion1
				0x00, 0x00, 0x00, 0x00, // FirmwareVersion2
				0x01,   // String Index
				1 << 3, // Characteristics
			},
			strings:  []string{"Test TPM"},
			complete: true,
			want: TPMDevice{
				VendorID:         [4]byte{0x00, 0x00, 0x00, 0x00},
				MajorSpecVersion: 2,
				MinorSpecVersion: 3,
				FirmwareVersion1: 0x00020001,
				FirmwareVersion2: 0x00000000,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristicsFamilyConfigurableViaFirmwareUpdate,
				OEMDefined:       0x12345678,
			},
			wantErr: fmt.Errorf("TPMDevice incomplete, got 8 of 9 fields"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := smbios.Table{
				Data:    tt.buffer,
				Strings: tt.strings,
			}
			TPMDev := &TPMDevice{
				Table: table,
			}

			// We need to modify tt.want with runtime data
			tt.want.Table = smbios.Table{
				Header: smbios.Header{
					Type:   smbios.TableType(tt.buffer[0]),
					Length: tt.buffer[1],
					Handle: binary.BigEndian.Uint16([]byte{tt.buffer[3], tt.buffer[2]}),
				},
				Data:    tt.buffer,
				Strings: tt.strings,
			}

			off, err := parseStruct(&table, 0, tt.complete, TPMDev)
			if err != tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("parseStruct() = %d, '%v' want '%v'", off, err, tt.wantErr)
				}
			}
			if tt.wantErr == nil {
				if TPMDev.VendorID != tt.want.VendorID {
					t.Errorf("parseStruct().VendorID = %q, want %q", TPMDev.VendorID, tt.want.VendorID)
				}

				if TPMDev.MajorSpecVersion != tt.want.MajorSpecVersion {
					t.Errorf("parseStruct().MajorSpecVersion = %q, want %q", TPMDev.MajorSpecVersion, tt.want.MajorSpecVersion)
				}

				if TPMDev.MinorSpecVersion != tt.want.MinorSpecVersion {
					t.Errorf("parseStruct().MinorSpecVersion = %q, want %q", TPMDev.MinorSpecVersion, tt.want.MinorSpecVersion)
				}

				if TPMDev.FirmwareVersion1 != tt.want.FirmwareVersion1 {
					t.Errorf("parseStruct().FirmwareVersion1 = %q, want %q", TPMDev.FirmwareVersion1, tt.want.FirmwareVersion1)
				}

				if TPMDev.FirmwareVersion2 != tt.want.FirmwareVersion2 {
					t.Errorf("parseStruct().FirmwareVersion2 = %q, want %q", TPMDev.FirmwareVersion2, tt.want.FirmwareVersion2)
				}

				if TPMDev.Description != tt.want.Description {
					t.Errorf("parseStruct().Description = %q, want %q", TPMDev.Description, tt.want.Description)
				}

				if TPMDev.Characteristics != tt.want.Characteristics {
					t.Errorf("parseStruct().Characteristics = %q, want %q", TPMDev.Characteristics, tt.want.Characteristics)
				}

				if TPMDev.OEMDefined != tt.want.OEMDefined {
					t.Errorf("parseStruct().OEMDefined = %q, want %q", TPMDev.OEMDefined, tt.want.OEMDefined)
				}
			}
		})
	}

}
