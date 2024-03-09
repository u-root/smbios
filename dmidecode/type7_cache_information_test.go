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

func TestCacheSizeBytes2Or1(t *testing.T) {
	tests := []struct {
		name  string
		size1 uint16
		size2 uint32
		want  uint64
	}{
		{
			name:  "No high bit set",
			size1: 0x1234,
			size2: 0x12345678,
			want:  0x48D159E000,
		},
		{
			name:  "High bit set",
			size1: 0x1234,
			size2: 0x80023456,
			want:  0x234560000,
		},
		{
			name:  "size2 zero",
			size1: 0x1234,
			size2: 0x80000000,
			want:  0x48D000,
		},
		{
			name:  "Zero",
			size1: 0x8000,
			size2: 0x80000000,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := cacheSizeBytes2Or1(tt.size1, tt.size2)
			if size != tt.want {
				t.Errorf("%q failed. Got: %q, Want: %q", tt.name, size, tt.want)
			}
		})
	}
}

func TestCacheInfoString(t *testing.T) {
	tests := []struct {
		name string
		val  CacheInfo
		want string
	}{
		{
			name: "Full details",
			val: CacheInfo{
				Header: smbios.Header{
					Length: 0x2a,
				},
				SocketDesignation:   "",
				Configuration:       0x03,
				MaximumSize:         0x100,
				InstalledSize:       0x3F,
				SupportedSRAMType:   CacheSRAMTypePipelineBurst,
				CurrentSRAMType:     CacheSRAMTypeOther,
				Speed:               0x4,
				ErrorCorrectionType: CacheErrorCorrectionTypeParity,
				SystemType:          CacheSystemTypeUnified,
				Associativity:       CacheAssociativity16waySetAssociative,
				MaximumSize2:        0x200,
				InstalledSize2:      0x00,
			},
			want: `Handle 0x0000, DMI type 0, 42 bytes
BIOS Information
	Socket Designation: Not Specified
	Configuration: Disabled, Not Socketed, Level 4
	Operational Mode: Write Through
	Location: Internal
	Installed Size: 63 kB
	Maximum Size: 512 kB
	Supported SRAM Types:
		Pipeline Burst
	Installed SRAM Type: Other
	Speed: 4 ns
	Error Correction Type: Parity
	System Type: Unified
	Associativity: 16-way Set-associative`,
		},
		{
			name: "More details",
			val: CacheInfo{
				Header: smbios.Header{
					Length: 0x2c,
				},
				SocketDesignation:   "",
				Configuration:       0x3A8,
				MaximumSize:         0x100,
				InstalledSize:       0x3F,
				SupportedSRAMType:   CacheSRAMTypePipelineBurst,
				CurrentSRAMType:     CacheSRAMTypeOther,
				Speed:               0x4,
				ErrorCorrectionType: CacheErrorCorrectionTypeParity,
				SystemType:          CacheSystemTypeUnified,
				Associativity:       CacheAssociativity16waySetAssociative,
				MaximumSize2:        0x200,
				InstalledSize2:      0x00,
			},
			want: `Handle 0x0000, DMI type 0, 44 bytes
BIOS Information
	Socket Designation: Not Specified
	Configuration: Enabled, Socketed, Level 1
	Operational Mode: Unknown
	Location: External
	Installed Size: 63 kB
	Maximum Size: 512 kB
	Supported SRAM Types:
		Pipeline Burst
	Installed SRAM Type: Other
	Speed: 4 ns
	Error Correction Type: Parity
	System Type: Unified
	Associativity: 16-way Set-associative`,
		},
		{
			name: "More details",
			val: CacheInfo{
				Header: smbios.Header{
					Length: 0x2c,
				},
				SocketDesignation:   "",
				Configuration:       0x2CA,
				MaximumSize:         0x100,
				InstalledSize:       0x3F,
				SupportedSRAMType:   CacheSRAMTypePipelineBurst,
				CurrentSRAMType:     CacheSRAMTypeOther,
				Speed:               0x4,
				ErrorCorrectionType: CacheErrorCorrectionTypeParity,
				SystemType:          CacheSystemTypeUnified,
				Associativity:       CacheAssociativity16waySetAssociative,
				MaximumSize2:        0x200,
				InstalledSize2:      0x00,
			},
			want: `Handle 0x0000, DMI type 0, 44 bytes
BIOS Information
	Socket Designation: Not Specified
	Configuration: Enabled, Socketed, Level 3
	Operational Mode: Varies With Memory Address
	Location: Reserved
	Installed Size: 63 kB
	Maximum Size: 512 kB
	Supported SRAM Types:
		Pipeline Burst
	Installed SRAM Type: Other
	Speed: 4 ns
	Error Correction Type: Parity
	System Type: Unified
	Associativity: 16-way Set-associative`,
		},
		{
			name: "More details",
			val: CacheInfo{
				Header: smbios.Header{
					Length: 0x2c,
				},
				SocketDesignation:   "",
				Configuration:       0x1EA,
				MaximumSize:         0x100,
				InstalledSize:       0x3F,
				SupportedSRAMType:   CacheSRAMTypePipelineBurst,
				CurrentSRAMType:     CacheSRAMTypeOther,
				Speed:               0x4,
				ErrorCorrectionType: CacheErrorCorrectionTypeParity,
				SystemType:          CacheSystemTypeUnified,
				Associativity:       CacheAssociativity16waySetAssociative,
				MaximumSize2:        0x200,
				InstalledSize2:      0x00,
			},
			want: `Handle 0x0000, DMI type 0, 44 bytes
BIOS Information
	Socket Designation: Not Specified
	Configuration: Enabled, Socketed, Level 3
	Operational Mode: Write Back
	Location: Unknown
	Installed Size: 63 kB
	Maximum Size: 512 kB
	Supported SRAM Types:
		Pipeline Burst
	Installed SRAM Type: Other
	Speed: 4 ns
	Error Correction Type: Parity
	System Type: Unified
	Associativity: 16-way Set-associative`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.String()
			if got != tt.want {
				t.Errorf("String = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInfoCache(t *testing.T) {
	for _, tt := range []struct {
		name  string
		table *smbios.Table
		want  *CacheInfo
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
					Type: smbios.TableTypeCacheInfo,
				},
			},
			err: io.ErrUnexpectedEOF,
		},

		{
			name: "Parse valid CacheInfo",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeCacheInfo,
				},
				Data: []byte{
					0x00,
					0x01, 0x02,
					0x03, 0x04,
					0x05, 0x06,
					0x07, 0x08,
					0x09, 0x0a,
				},
			},
			want: &CacheInfo{
				Header: smbios.Header{
					Type: smbios.TableTypeCacheInfo,
				},
				Configuration:     0x0201,
				MaximumSize:       0x0403,
				InstalledSize:     0x0605,
				SupportedSRAMType: 0x0807,
				CurrentSRAMType:   0x0a09,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCacheInfo(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseCacheInfo = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCacheInfo = %v, want %v", got, tt.want)
			}
		})
	}
}
