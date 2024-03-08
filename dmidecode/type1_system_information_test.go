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

func TestSystemInfoString(t *testing.T) {
	tests := []struct {
		name string
		val  SystemInfo
		want string
	}{
		{
			name: "All Infos provided",
			val: SystemInfo{
				Header: smbios.Header{
					Length: 26,
				},
				Manufacturer: "u-root testing",
				ProductName:  "Illusion",
				Version:      "1.0",
				SerialNumber: "UR00T1234",
				UUID:         UUID{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
				SKUNumber:    "3a",
				Family:       "UR00T1234",
			},
			want: `Handle 0x0000, DMI type 0, 26 bytes
BIOS Information
	Manufacturer: u-root testing
	Product Name: Illusion
	Version: 1.0
	Serial Number: UR00T1234
	UUID: 03020100-0504-0706-0809-0a0b0c0d0e0f
	Wake-up Type: Reserved
	SKU Number: 3a
	Family: UR00T1234`,
		},
		{
			name: "UUID not present",
			val: SystemInfo{
				Header: smbios.Header{
					Length: 8,
				},
				Manufacturer: "u-root testing",
				ProductName:  "Illusion",
				Version:      "1.0",
				SerialNumber: "UR00T1234",
				UUID:         UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SKUNumber:    "3a",
				Family:       "UR00T1234",
			},
			want: `Handle 0x0000, DMI type 0, 8 bytes
BIOS Information
	Manufacturer: u-root testing
	Product Name: Illusion
	Version: 1.0
	Serial Number: UR00T1234
	UUID: Not Present
	Wake-up Type: Reserved`,
		},
		{
			name: "UUID not settable",
			val: SystemInfo{
				Header: smbios.Header{
					Length: 8,
				},
				Manufacturer: "u-root testing",
				ProductName:  "Illusion",
				Version:      "1.0",
				SerialNumber: "UR00T1234",
				UUID:         UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				SKUNumber:    "3a",
				Family:       "UR00T1234",
			},
			want: `Handle 0x0000, DMI type 0, 8 bytes
BIOS Information
	Manufacturer: u-root testing
	Product Name: Illusion
	Version: 1.0
	Serial Number: UR00T1234
	UUID: Not Settable
	Wake-up Type: Reserved`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.val.String()
			if result != tt.want {
				t.Errorf("SystemInfo().String(): '%s', want '%s'", result, tt.want)
			}
		})
	}
}

func TestUUIDParseField(t *testing.T) {
	tests := []struct {
		name string
		val  smbios.Table
		want string
	}{
		{
			name: "Valid UUID",
			val: smbios.Table{
				Data: []byte{0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03,
					0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03},
			},
			want: "03020100-0100-0302-0001-020300010203",
		},
	}

	for _, tt := range tests {
		uuid := UUID([16]byte{})
		_, err := uuid.ParseField(&tt.val, 0)
		if err != nil {
			t.Errorf("ParseField(): '%v', want nil", err)
		}
		if uuid.String() != tt.want {
			t.Errorf("ParseField(): '%s', want '%s'", uuid.String(), tt.want)
		}
	}
}

func TestParseSystemInfo(t *testing.T) {
	for _, tt := range []struct {
		table *smbios.Table
		want  *SystemInfo
		err   error
	}{
		{
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeBIOSInfo,
				},
			},
			err: ErrUnexpectedTableType,
		},
		{
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeSystemInfo,
				},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			table: &smbios.Table{
				Header: smbios.Header{
					Length: 4 + 4 + 16 + 2,
					Type:   smbios.TableTypeSystemInfo,
				},
				Data: []byte{
					0x00, 0x01, 0x00, 0x00,
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
					0x00,
					0x00,
					0x00,
				},
				Strings: []string{
					"foobar",
				},
			},
			want: &SystemInfo{
				Header: smbios.Header{
					Length: 4 + 4 + 16 + 2,
					Type:   smbios.TableTypeSystemInfo,
				},
				ProductName: "foobar",
				UUID: UUID{
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				},
			},
		},
		/*
			TODO: deal with BAD INDEX cases, check other SMBIOS implementations.
			{
				table: &smbios.Table{
					Header: smbios.Header{
						Type: smbios.TableTypeSystemInfo,
					},
					Data: []byte{
						0x00, 0x01, 0x02, 0x03,
					},
				},
				err: io.ErrUnexpectedEOF,
			},
		*/
	} {
		t.Run("", func(t *testing.T) {
			got, err := ParseSystemInfo(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseSystemInfo = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSystemInfo = %v, want %v", got, tt.want)
			}
		})
	}
}
