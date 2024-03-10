// Copyright 2016-2022 the u-root Authors. All rights reserved
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

func TestParseSystemSlots(t *testing.T) {
	for _, tt := range []struct {
		name  string
		want  *SystemSlots
		table *smbios.Table
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
					Type: smbios.TableTypeSystemSlots,
				},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "Parse valid SystemSlots",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeSystemSlots,
				},
				Data: []byte{
					0x00,
					0x00,
					0x00,
					0x00,
					0x00,
					0x01, 0x02,
					0x00,
					0x00,
					0x03, 0x02,
					0x00,
					0x00,
					0x00,
				},
			},
			want: &SystemSlots{
				Table: smbios.Table{
					Header: smbios.Header{
						Type: smbios.TableTypeSystemSlots,
					},
					Data: []byte{
						0x00,
						0x00,
						0x00,
						0x00,
						0x00,
						0x01, 0x02,
						0x00,
						0x00,
						0x03, 0x02,
						0x00,
						0x00,
						0x00,
					},
				},
				SlotID:             0x0201,
				SegmentGroupNumber: 0x0203,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSystemSlots(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseSystemSlots = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSystemSlots =\n%#v, want\n%#v", got, tt.want)
			}
		})
	}
}
