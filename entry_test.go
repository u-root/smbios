// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package smbios

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestParseEntry(t *testing.T) {
	for _, tt := range []struct {
		b    []byte
		want EntryPoint
		err  error
	}{
		{
			b: []byte{
				'_', 'S', 'M', '_',
				0x73, // checksum
				31,
				1, 1,
				14, 0,
				0,
				0, 0, 0, 0, 0,
				'_', 'D', 'M', 'I', '_',
				0x68, // checksum
				0, 0,
				0, 0, 0, 0,
				0, 0,
				0,
			},
			want: &Entry32{
				Anchor:            [4]byte{95, 83, 77, 95},
				Checksum:          0x73,
				Length:            0x1F,
				MajorVersion:      0x01,
				MinorVersion:      0x01,
				StructMaxSize:     0x000E,
				Revision:          0x00,
				Reserved:          [5]byte{0x00, 0x00, 0x00, 0x00, 0x00},
				IntAnchor:         [5]byte{95, 68, 77, 73, 95},
				IntChecksum:       0x68,
				StructTableLength: 0x0000,
				StructTableAddr:   0x00000000,
				NumberOfStructs:   0x0000,
				BCDRevision:       0x00,
			},
		},
		{
			b: []byte{
				'_', 'S', '_', 'M',
				0x00,
			},
			err: ErrInvalidAnchor,
		},
		{
			b: []byte{
				'_', 'S', '_', 'M',
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				'_', 'S', 'M', '_',
				0x73, // checksum
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				'_', 'S', 'M', '3', '_',
				0x73, // checksum
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				'_', 'S', 'M', '_',
				0x73, // checksum
				31,
				1, 1,
				14, 0,
				0,
				0, 0, 0, 0, 0,
				'_', 'D', 'M', 'I', '_',
				0x68, // checksum
				0, 0,
				0, 0, 0, 0,
				0, 0,
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				'_', 'S', 'M', '3', '_',
				0x5F,
				0x18,
				2, 1, 1,
				0,
				0,
				255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255,
			},
			want: &Entry64{
				Anchor:          [5]byte{95, 83, 77, 51, 95},
				Checksum:        0x5F,
				Length:          0x18,
				MajorVersion:    0x02,
				MinorVersion:    0x01,
				DocRev:          0x01,
				Revision:        0x00,
				Reserved:        0x00,
				StructMaxSize:   0xFFFFFFFF,
				StructTableAddr: 0xFFFFFFFFFFFFFFFF,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := ParseEntry(bytes.NewReader(tt.b))
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseEntry = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEntry =\n%v, want\n%v", got, tt.want)
			}
		})
	}
}

func TestMarshalEntry(t *testing.T) {
	for _, tt := range []struct {
		want []byte
		e    EntryPoint
		err  error
	}{
		{
			want: []byte{
				'_', 'S', 'M', '_',
				0x73, // checksum
				31,
				1, 1,
				14, 0,
				0,
				0, 0, 0, 0, 0,
				'_', 'D', 'M', 'I', '_',
				0x68, // checksum
				0, 0,
				0, 0, 0, 0,
				0, 0,
				0,
			},
			e: &Entry32{
				Anchor:            [4]byte{95, 83, 77, 95},
				Checksum:          0x73,
				Length:            0x1F,
				MajorVersion:      0x01,
				MinorVersion:      0x01,
				StructMaxSize:     0x000E,
				Revision:          0x00,
				Reserved:          [5]byte{0x00, 0x00, 0x00, 0x00, 0x00},
				IntAnchor:         [5]byte{95, 68, 77, 73, 95},
				IntChecksum:       0x68,
				StructTableLength: 0x0000,
				StructTableAddr:   0x00000000,
				NumberOfStructs:   0x0000,
				BCDRevision:       0x00,
			},
		},
		{
			want: []byte{
				'_', 'S', 'M', '3', '_',
				0x5F,
				0x18,
				2, 1, 1,
				0,
				0,
				255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255,
			},
			e: &Entry64{
				Anchor:          [5]byte{95, 83, 77, 51, 95},
				Checksum:        0x5F,
				Length:          0x18,
				MajorVersion:    0x02,
				MinorVersion:    0x01,
				DocRev:          0x01,
				Revision:        0x00,
				Reserved:        0x00,
				StructMaxSize:   0xFFFFFFFF,
				StructTableAddr: 0xFFFFFFFFFFFFFFFF,
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := tt.e.MarshalBinary()
			if !errors.Is(err, tt.err) {
				t.Errorf("Marshal = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal =\n%v, want\n%v", got, tt.want)
			}
		})
	}
}
