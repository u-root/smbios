// Copyright 2016-2022 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"fmt"
	"testing"

	"github.com/u-root/smbios"
)

func TestParseSystemSlots(t *testing.T) {
	tests := []struct {
		name  string
		val   SystemSlots
		table smbios.Table
		want  error
	}{
		{
			name: "Invalid Type",
			val:  SystemSlots{},
			table: smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeBIOSInfo,
				},
				Data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
					0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
					0x1a},
			},
			want: fmt.Errorf("invalid table type 0"),
		},
		{
			name: "Required fields are missing",
			val:  SystemSlots{},
			table: smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeSystemSlots,
				},
				Data: []byte{},
			},
			want: fmt.Errorf("required fields missing"),
		},
		{
			name: "Error parsing structure",
			val:  SystemSlots{},
			table: smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeSystemSlots,
				},
				Data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
					0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
					0x1a},
			},
			want: fmt.Errorf("error parsing structure"),
		},
		{
			name: "Parse valid SystemSlots",
			val:  SystemSlots{},
			table: smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeSystemSlots,
				},
				Data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
					0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
					0x1a},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseStruct := func(t *smbios.Table, off int, complete bool, sp interface{}) (int, error) {
				return 0, tt.want
			}
			_, err := parseSystemSlots(parseStruct, &tt.table)

			if !checkError(err, tt.want) {
				t.Errorf("%q failed. Got: %q, Want: %q", tt.name, err, tt.want)
			}
		})
	}
}
