// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"fmt"
	"io"
	"strings"

	"github.com/u-root/smbios"
)

// CacheInfo is defined in DSP0134 7.8.
type CacheInfo struct {
	smbios.Header       `smbios:"-"`
	SocketDesignation   string                   // 04h
	Configuration       uint16                   // 05h
	MaximumSize         uint16                   // 07h
	InstalledSize       uint16                   // 09h
	SupportedSRAMType   CacheSRAMType            // 0Bh
	CurrentSRAMType     CacheSRAMType            // 0Dh
	Speed               uint8                    // 0Fh
	ErrorCorrectionType CacheErrorCorrectionType // 10h
	SystemType          CacheSystemType          // 11h
	Associativity       CacheAssociativity       // 12h
	MaximumSize2        uint32                   // 13h
	InstalledSize2      uint32                   // 17h
}

// ParseCacheInfo parses a generic smbios.Table into CacheInfo.
func ParseCacheInfo(t *smbios.Table) (*CacheInfo, error) {
	if t.Type != smbios.TableTypeCacheInfo {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	if t.Len() < 0xf {
		return nil, fmt.Errorf("%w: cache info table must be at least %d bytes", io.ErrUnexpectedEOF, 0xf)
	}
	ci := &CacheInfo{Header: t.Header}
	_, err := parseStruct(t, 0 /* off */, false /* complete */, ci)
	if err != nil {
		return nil, err
	}
	return ci, nil
}

func cacheSizeBytes2Or1(size1 uint16, size2 uint32) uint64 {
	mul2 := uint64(1024)
	if size2&0x80000000 != 0 {
		mul2 *= 64
	}
	if size2Bytes := uint64(size2&0x7fffffff) * mul2; size2Bytes != 0 {
		return size2Bytes
	}
	mul1 := uint64(1024)
	if size1&0x8000 != 0 {
		mul1 *= 64
	}
	return uint64(size1&0x7fff) * mul1
}

// GetMaxSizeBytes returns the maximum size  of the cache that can be installed, in bytes.
func (ci *CacheInfo) GetMaxSizeBytes() uint64 {
	return cacheSizeBytes2Or1(ci.MaximumSize, ci.MaximumSize2)
}

// GetInstalledSizeBytes returns the currently installed cache size, in bytes.
func (ci *CacheInfo) GetInstalledSizeBytes() uint64 {
	return cacheSizeBytes2Or1(ci.InstalledSize, ci.InstalledSize2)
}

func (ci *CacheInfo) String() string {
	enDis := "Disabled"
	if ci.Configuration&0x80 != 0 {
		enDis = "Enabled"
	}
	sock := "Not Socketed"
	if ci.Configuration&0x8 != 0 {
		sock = "Socketed"
	}

	om := ""
	switch (ci.Configuration >> 8) & 3 {
	case 0:
		om = "Write Through"
	case 1:
		om = "Write Back"
	case 2:
		om = "Varies With Memory Address"
	case 3:
		om = "Unknown"
	}

	loc := ""
	switch (ci.Configuration >> 5) & 3 {
	case 0:
		loc = "Internal"
	case 1:
		loc = "External"
	case 2:
		loc = "Reserved"
	case 3:
		loc = "Unknown"
	}

	speedStr := "Unknown"
	if ci.Speed > 0 {
		speedStr = fmt.Sprintf("%d ns", ci.Speed)
	}

	lines := []string{
		ci.Header.String(),
		fmt.Sprintf("Socket Designation: %s", smbiosStr(ci.SocketDesignation)),
		fmt.Sprintf("Configuration: %s, %s, Level %d", enDis, sock, (ci.Configuration&7)+1),
		fmt.Sprintf("Operational Mode: %s", om),
		fmt.Sprintf("Location: %s", loc),
		fmt.Sprintf("Installed Size: %s", kmgt(ci.GetInstalledSizeBytes())),
		fmt.Sprintf("Maximum Size: %s", kmgt(ci.GetMaxSizeBytes())),
		fmt.Sprintf("Supported SRAM Types:\n%s", ci.SupportedSRAMType),
		fmt.Sprintf("Installed SRAM Type: %s", strings.TrimSpace(ci.CurrentSRAMType.String())),
	}
	if ci.Length > 0xf {
		lines = append(lines,
			fmt.Sprintf("Speed: %s", speedStr),
			fmt.Sprintf("Error Correction Type: %s", ci.ErrorCorrectionType),
			fmt.Sprintf("System Type: %s", ci.SystemType),
			fmt.Sprintf("Associativity: %s", ci.Associativity),
		)
	}
	return strings.Join(lines, "\n\t")
}

// CacheSRAMType is defined in DSP0134 7.8.2.
type CacheSRAMType uint16

// CacheSRAMType fields are defined in DSP0134 7.8.2.
const (
	CacheSRAMTypeOther         CacheSRAMType = 1 << 0 // Other
	CacheSRAMTypeUnknown       CacheSRAMType = 1 << 1 // Unknown
	CacheSRAMTypeNonBurst      CacheSRAMType = 1 << 2 // Non-Burst
	CacheSRAMTypeBurst         CacheSRAMType = 1 << 3 // Burst
	CacheSRAMTypePipelineBurst CacheSRAMType = 1 << 4 // Pipeline Burst
	CacheSRAMTypeSynchronous   CacheSRAMType = 1 << 5 // Synchronous
	CacheSRAMTypeAsynchronous  CacheSRAMType = 1 << 6 // Asynchronous
)

var cacheSRAM = map[CacheSRAMType]string{
	CacheSRAMTypeOther:         "Other",
	CacheSRAMTypeUnknown:       "Unknown",
	CacheSRAMTypeNonBurst:      "Non-Burst",
	CacheSRAMTypeBurst:         "Burst",
	CacheSRAMTypePipelineBurst: "Pipeline Burst",
	CacheSRAMTypeSynchronous:   "Synchronous",
	CacheSRAMTypeAsynchronous:  "Asynchronous",
}

func (v CacheSRAMType) String() string {
	var lines []string
	for i := 0; i < 7; i++ {
		if v&(1<<i) != 0 {
			lines = append(lines, cacheSRAM[1<<i])
		}
	}
	return "\t\t" + strings.Join(lines, "\n\t\t")
}

// CacheErrorCorrectionType is defined in DSP0134 7.8.3.
type CacheErrorCorrectionType uint8

// CacheErrorCorrectionType values are defined in DSP0134 7.8.3.
const (
	CacheErrorCorrectionTypeOther        CacheErrorCorrectionType = 0x01 // Other
	CacheErrorCorrectionTypeUnknown      CacheErrorCorrectionType = 0x02 // Unknown
	CacheErrorCorrectionTypeNone         CacheErrorCorrectionType = 0x03 // None
	CacheErrorCorrectionTypeParity       CacheErrorCorrectionType = 0x04 // Parity
	CacheErrorCorrectionTypeSinglebitECC CacheErrorCorrectionType = 0x05 // Single-bit ECC
	CacheErrorCorrectionTypeMultibitECC  CacheErrorCorrectionType = 0x06 // Multi-bit ECC
)

var cacheECStr = map[CacheErrorCorrectionType]string{
	CacheErrorCorrectionTypeOther:        "Other",
	CacheErrorCorrectionTypeUnknown:      "Unknown",
	CacheErrorCorrectionTypeNone:         "None",
	CacheErrorCorrectionTypeParity:       "Parity",
	CacheErrorCorrectionTypeSinglebitECC: "Single-bit ECC",
	CacheErrorCorrectionTypeMultibitECC:  "Multi-bit ECC",
}

func (v CacheErrorCorrectionType) String() string {
	if name, ok := cacheECStr[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// CacheSystemType is defined in DSP0134 7.8.4.
type CacheSystemType uint8

// CacheSystemType values are defined in DSP0134 7.8.4.
const (
	CacheSystemTypeOther       CacheSystemType = 0x01 // Other
	CacheSystemTypeUnknown     CacheSystemType = 0x02 // Unknown
	CacheSystemTypeInstruction CacheSystemType = 0x03 // Instruction
	CacheSystemTypeData        CacheSystemType = 0x04 // Data
	CacheSystemTypeUnified     CacheSystemType = 0x05 // Unified
)

var cacheSTStr = map[CacheSystemType]string{
	CacheSystemTypeOther:       "Other",
	CacheSystemTypeUnknown:     "Unknown",
	CacheSystemTypeInstruction: "Instruction",
	CacheSystemTypeData:        "Data",
	CacheSystemTypeUnified:     "Unified",
}

func (v CacheSystemType) String() string {
	if name, ok := cacheSTStr[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// CacheAssociativity is defined in DSP0134 7.8.5.
type CacheAssociativity uint8

// CacheAssociativity values are defined in DSP0134 7.8.5.
const (
	CacheAssociativityOther               CacheAssociativity = 0x01 // Other
	CacheAssociativityUnknown             CacheAssociativity = 0x02 // Unknown
	CacheAssociativityDirectMapped        CacheAssociativity = 0x03 // Direct Mapped
	CacheAssociativity2waySetAssociative  CacheAssociativity = 0x04 // 2-way Set-associative
	CacheAssociativity4waySetAssociative  CacheAssociativity = 0x05 // 4-way Set-associative
	CacheAssociativityFullyAssociative    CacheAssociativity = 0x06 // Fully Associative
	CacheAssociativity8waySetAssociative  CacheAssociativity = 0x07 // 8-way Set-associative
	CacheAssociativity16waySetAssociative CacheAssociativity = 0x08 // 16-way Set-associative
	CacheAssociativity12waySetAssociative CacheAssociativity = 0x09 // 12-way Set-associative
	CacheAssociativity24waySetAssociative CacheAssociativity = 0x0a // 24-way Set-associative
	CacheAssociativity32waySetAssociative CacheAssociativity = 0x0b // 32-way Set-associative
	CacheAssociativity48waySetAssociative CacheAssociativity = 0x0c // 48-way Set-associative
	CacheAssociativity64waySetAssociative CacheAssociativity = 0x0d // 64-way Set-associative
	CacheAssociativity20waySetAssociative CacheAssociativity = 0x0e // 20-way Set-associative
)

var cacheAssocStr = map[CacheAssociativity]string{
	CacheAssociativityOther:               "Other",
	CacheAssociativityUnknown:             "Unknown",
	CacheAssociativityDirectMapped:        "Direct Mapped",
	CacheAssociativity2waySetAssociative:  "2-way Set-associative",
	CacheAssociativity4waySetAssociative:  "4-way Set-associative",
	CacheAssociativityFullyAssociative:    "Fully Associative",
	CacheAssociativity8waySetAssociative:  "8-way Set-associative",
	CacheAssociativity16waySetAssociative: "16-way Set-associative",
	CacheAssociativity12waySetAssociative: "12-way Set-associative",
	CacheAssociativity24waySetAssociative: "24-way Set-associative",
	CacheAssociativity32waySetAssociative: "32-way Set-associative",
	CacheAssociativity48waySetAssociative: "48-way Set-associative",
	CacheAssociativity64waySetAssociative: "64-way Set-associative",
	CacheAssociativity20waySetAssociative: "20-way Set-associative",
}

func (v CacheAssociativity) String() string {
	if name, ok := cacheAssocStr[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}
