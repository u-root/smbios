// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/u-root/smbios"
)

// BaseboardInfo is defined in DSP0134 7.3.
type BaseboardInfo struct {
	smbios.Header     `smbios:"-"`
	Manufacturer      string        // 04h
	Product           string        // 05h
	Version           string        // 06h
	SerialNumber      string        // 07h
	AssetTag          string        // 08h
	BoardFeatures     BoardFeatures // 09h
	LocationInChassis string        // 0Ah
	ChassisHandle     uint16        // 0Bh
	BoardType         BoardType     // 0Dh
	ObjectHandles     ObjectHandles // 0Eh
}

// ParseBaseboardInfo parses a generic smbios.Table into BaseboardInfo.
func ParseBaseboardInfo(t *smbios.Table) (*BaseboardInfo, error) {
	if t.Type != smbios.TableTypeBaseboardInfo {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	// Defined in DSP0134 7.3, length of the structure is at least 08h.
	if t.Len() < 0x8 {
		return nil, fmt.Errorf("%w: baseboard info table must be at least %d bytes", io.ErrUnexpectedEOF, 8)
	}
	bi := &BaseboardInfo{Header: t.Header}
	_, err := parseStruct(t, 0 /* off */, false /* complete */, bi)
	if err != nil {
		return nil, err
	}
	return bi, nil
}

// Typ implements Table.Typ.
func (bi BaseboardInfo) Typ() smbios.TableType {
	return smbios.TableTypeBaseboardInfo
}

func (bi *BaseboardInfo) String() string {
	lines := []string{
		bi.Header.String(),
		fmt.Sprintf("Manufacturer: %s", smbiosStr(bi.Manufacturer)),
		fmt.Sprintf("Product Name: %s", smbiosStr(bi.Product)),
		fmt.Sprintf("Version: %s", smbiosStr(bi.Version)),
		fmt.Sprintf("Serial Number: %s", smbiosStr(bi.SerialNumber)),
		fmt.Sprintf("Asset Tag: %s", smbiosStr(bi.AssetTag)),
		fmt.Sprintf("Features:\n%s", bi.BoardFeatures),
		fmt.Sprintf("Location In Chassis: %s", smbiosStr(bi.LocationInChassis)),
		fmt.Sprintf("Chassis Handle: 0x%04X", bi.ChassisHandle),
		fmt.Sprintf("Type: %s", bi.BoardType),
		bi.ObjectHandles.str(),
	}
	return strings.Join(lines, "\n\t")
}

// BoardFeatures is defined in DSP0134 7.3.1.
type BoardFeatures uint8

// BoardFeatures fields are defined in DSP0134 7.3.1.
const (
	BoardFeaturesIsHotSwappable                  BoardFeatures = 1 << 4 // Set to 1 if the board is hot swappable
	BoardFeaturesIsReplaceable                   BoardFeatures = 1 << 3 // Set to 1 if the board is replaceable
	BoardFeaturesIsRemovable                     BoardFeatures = 1 << 2 // Set to 1 if the board is removable
	BoardFeaturesRequiresAtLeastOneDaughterBoard BoardFeatures = 1 << 1 // Set to 1 if the board requires at least one daughter board or auxiliary card to function
	BoardFeaturesIsAHostingBoard                 BoardFeatures = 1 << 0 // Set to 1 if the board is a hosting board (for example, a motherboard)
)

var boardFeatureStr = map[BoardFeatures]string{
	BoardFeaturesIsAHostingBoard:                 "Board is a hosting board",
	BoardFeaturesRequiresAtLeastOneDaughterBoard: "Board requires at least one daughter board",
	BoardFeaturesIsRemovable:                     "Board is removable",
	BoardFeaturesIsReplaceable:                   "Board is replaceable",
	BoardFeaturesIsHotSwappable:                  "Board is hot swappable",
}

func (v BoardFeatures) String() string {
	var lines []string
	for i := 0; i < 5; i++ {
		if v&(1<<i) != 0 {
			lines = append(lines, boardFeatureStr[1<<i])
		}
	}
	return "\t\t" + strings.Join(lines, "\n\t\t")
}

// BoardType is defined in DSP0134 7.3.2.
type BoardType uint8

// BoardType values are defined in DSP0134 7.3.2.
const (
	BoardTypeUnknown                                 BoardType = 0x01 // Unknown
	BoardTypeOther                                   BoardType = 0x02 // Other
	BoardTypeServerBlade                             BoardType = 0x03 // Server Blade
	BoardTypeConnectivitySwitch                      BoardType = 0x04 // Connectivity Switch
	BoardTypeSystemManagementModule                  BoardType = 0x05 // System Management Module
	BoardTypeProcessorModule                         BoardType = 0x06 // Processor Module
	BoardTypeIOModule                                BoardType = 0x07 // I/O Module
	BoardTypeMemoryModule                            BoardType = 0x08 // Memory Module
	BoardTypeDaughterBoard                           BoardType = 0x09 // Daughter board
	BoardTypeMotherboardIncludesProcessorMemoryAndIO BoardType = 0x0a // Motherboard (includes processor, memory, and I/O)
	BoardTypeProcessorMemoryModule                   BoardType = 0x0b // Processor/Memory Module
	BoardTypeProcessorIOModule                       BoardType = 0x0c // Processor/IO Module
	BoardTypeInterconnectBoard                       BoardType = 0x0d // Interconnect board
)

var boardStrings = map[BoardType]string{
	BoardTypeUnknown:                                 "Unknown",
	BoardTypeOther:                                   "Other",
	BoardTypeServerBlade:                             "Server Blade",
	BoardTypeConnectivitySwitch:                      "Connectivity Switch",
	BoardTypeSystemManagementModule:                  "System Management Module",
	BoardTypeProcessorModule:                         "Processor Module",
	BoardTypeIOModule:                                "I/O Module",
	BoardTypeMemoryModule:                            "Memory Module",
	BoardTypeDaughterBoard:                           "Daughter board",
	BoardTypeMotherboardIncludesProcessorMemoryAndIO: "Motherboard",
	BoardTypeProcessorMemoryModule:                   "Processor/Memory Module",
	BoardTypeProcessorIOModule:                       "Processor/IO Module",
	BoardTypeInterconnectBoard:                       "Interconnect board",
}

func (v BoardType) String() string {
	if name, ok := boardStrings[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// ObjectHandles are defined in DSP0134 v4.7 Section 7.3 and embedded in the Baseboard structure.
type ObjectHandles []uint16

func (oh ObjectHandles) String() string {
	lines := []string{fmt.Sprintf("Contained Object Handles: %d", len(oh))}
	for _, h := range oh {
		lines = append(lines, fmt.Sprintf("\t0x%04X", h))
	}
	return strings.Join(lines, "\n")
}

func (oh ObjectHandles) str() string {
	lines := []string{fmt.Sprintf("Contained Object Handles: %d", len(oh))}
	for _, h := range oh {
		lines = append(lines, fmt.Sprintf("\t\t0x%04X", h))
	}
	return strings.Join(lines, "\n")
}

// WriteField writes a object handles as defined by DSP0134 Section 7.3.
func (oh *ObjectHandles) WriteField(t *smbios.Table) (int, error) {
	num := len(*oh)
	if num > math.MaxUint8 {
		return 0, fmt.Errorf("%w: too many object handles defined, can be maximum of 256", ErrInvalidArg)
	}
	t.WriteByte(uint8(num))
	for _, handle := range *oh {
		t.WriteWord(handle)
	}
	return 1 + 2*num, nil
}

// ParseField parses object handles as defined by DSP0134 Section 7.3.
func (oh *ObjectHandles) ParseField(t *smbios.Table, off int) (int, error) {
	num, err := t.GetByteAt(off)
	if err != nil {
		return off, err
	}
	off++

	for i := uint8(0); i < num; i++ {
		h, err := t.GetWordAt(off)
		if err != nil {
			return off, err
		}
		*oh = append(*oh, h)
		off += 2
	}
	return off, nil
}
