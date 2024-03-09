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

// BIOSInfo is defined in DSP0134 7.1.
type BIOSInfo struct {
	smbios.Header          `smbios:"-"`
	Vendor                 string        // 04h
	Version                string        // 05h
	StartingAddressSegment uint16        // 06h
	ReleaseDate            string        // 08h
	ROMSize                uint8         // 09h
	Characteristics        BIOSChars     // 0Ah
	CharacteristicsExt1    BIOSCharsExt1 // 12h
	CharacteristicsExt2    BIOSCharsExt2 // 13h
	BIOSMajor              uint8         `smbios:"default=0xff"` // 14h
	BIOSMinor              uint8         `smbios:"default=0xff"` // 15h
	ECMajor                uint8         `smbios:"default=0xff"` // 16h
	ECMinor                uint8         `smbios:"default=0xff"` // 17h
	ExtendedROMSize        uint16        // 18h
}

// ParseBIOSInfo parses a generic Table into BIOSInfo.
func ParseBIOSInfo(t *smbios.Table) (*BIOSInfo, error) {
	if t.Type != smbios.TableTypeBIOSInfo {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	if t.Len() < 0x12 {
		return nil, fmt.Errorf("%w: BIOS info table must be at least %d bytes", io.ErrUnexpectedEOF, 0x12)
	}
	bi := &BIOSInfo{
		Header: t.Header,
	}
	if _, err := parseStruct(t, 0 /* off */, false /* complete */, bi); err != nil {
		return nil, err
	}
	return bi, nil
}

// ROMSizeBytes returns ROM size in bytes.
func (bi *BIOSInfo) ROMSizeBytes() uint64 {
	if bi.ROMSize != 0xff || bi.ExtendedROMSize == 0 {
		return 65536 * (uint64(bi.ROMSize) + 1)
	}

	extSize := uint64(bi.ExtendedROMSize)
	unit := (extSize >> 14)

	multiplier := uint64(1)
	switch unit {
	case 0:
		multiplier = 1024 * 1024
	case 1:
		multiplier = 1024 * 1024 * 1024
	}
	return (extSize & 0x3fff) * multiplier
}

func (bi *BIOSInfo) String() string {
	lines := []string{
		bi.Header.String(),
		fmt.Sprintf("\tVendor: %s", smbiosStr(bi.Vendor)),
		fmt.Sprintf("\tVersion: %s", smbiosStr(bi.Version)),
		fmt.Sprintf("\tRelease Date: %s", smbiosStr(bi.ReleaseDate)),
	}
	if bi.StartingAddressSegment != 0 {
		lines = append(lines,
			fmt.Sprintf("\tAddress: 0x%04X0", bi.StartingAddressSegment),
			fmt.Sprintf("\tRuntime Size: %s", kmgt(uint64((0x10000-int(bi.StartingAddressSegment))<<4))),
		)
	}
	lines = append(lines,
		fmt.Sprintf("\tROM Size: %s", kmgt(bi.ROMSizeBytes())),
		fmt.Sprintf("\tCharacteristics:\n%s", bi.Characteristics),
		bi.CharacteristicsExt1.String(),
		bi.CharacteristicsExt2.String(),
	)
	if bi.BIOSMajor != 0xff { // 2.4+
		lines = append(lines, fmt.Sprintf("\tBIOS Revision: %d.%d", bi.BIOSMajor, bi.BIOSMinor))
	}
	if bi.ECMajor != 0xff {
		lines = append(lines, fmt.Sprintf("\tFirmware Revision: %d.%d", bi.ECMajor, bi.ECMinor))
	}
	return strings.Join(lines, "\n")
}

// BIOSChars are BIOS characteristics as defined in DSP0134 7.1.1.
type BIOSChars uint64

// BIOSChars fields are defined in DSP0134 7.1.1.
const (
	// Reserved.
	BIOSCharsReserved BIOSChars = 1 << 0
	// Reserved.
	BIOSCharsReserved2 BIOSChars = 1 << 1
	// Unknown.
	BIOSCharsUnknown BIOSChars = 1 << 2
	// BIOS Characteristics are not supported.
	BIOSCharsAreNotSupported BIOSChars = 1 << 3
	// ISA is supported.
	BIOSCharsISA BIOSChars = 1 << 4
	// MCA is supported.
	BIOSCharsMCA BIOSChars = 1 << 5
	// EISA is supported.
	BIOSCharsEISA BIOSChars = 1 << 6
	// PCI is supported.
	BIOSCharsPCI BIOSChars = 1 << 7
	// PC card (PCMCIA) is supported.
	BIOSCharsPCMCIA BIOSChars = 1 << 8
	// Plug and Play is supported.
	BIOSCharsPlugAndPlay BIOSChars = 1 << 9
	// APM is supported.
	BIOSCharsAPM BIOSChars = 1 << 10
	// BIOS is upgradeable (Flash).
	BIOSCharsBIOSUpgradeableFlash BIOSChars = 1 << 11
	// BIOS shadowing is allowed.
	BIOSCharsBIOSShadowingIsAllowed BIOSChars = 1 << 12
	// VL-VESA is supported.
	BIOSCharsVLVESA BIOSChars = 1 << 13
	// ESCD support is available.
	BIOSCharsESCD BIOSChars = 1 << 14
	// Boot from CD is supported.
	BIOSCharsBootFromCD BIOSChars = 1 << 15
	// Selectable boot is supported.
	BIOSCharsSelectableBoot BIOSChars = 1 << 16
	// BIOS ROM is socketed.
	BIOSCharsBIOSROMSocketed BIOSChars = 1 << 17
	// Boot from PC card (PCMCIA) is supported.
	BIOSCharsBootFromPCMCIA BIOSChars = 1 << 18
	// EDD specification is supported.
	BIOSCharsEDD BIOSChars = 1 << 19
	// Japanese floppy for NEC 9800 1.2 MB (3.5”, 1K bytes/sector, 360 RPM) is supported.
	BIOSCharsJapaneseFloppyNEC BIOSChars = 1 << 20
	// Japanese floppy for Toshiba 1.2 MB (3.5”, 360 RPM) is supported.
	BIOSCharsJapaneseFloppyToshiba BIOSChars = 1 << 21
	// 5.25” / 360 KB floppy services are supported.
	BIOSChars360KBFloppy BIOSChars = 1 << 22
	// 5.25” /1.2 MB floppy services are supported.
	BIOSChars12MBFloppy BIOSChars = 1 << 23
	// 3.5” / 720 KB floppy services are supported.
	BIOSChars720KBFloppy BIOSChars = 1 << 24
	// 3.5” / 2.88 MB floppy services are supported.
	BIOSChars288MBFloppy BIOSChars = 1 << 25
	// Int 5h, print screen Service is supported.
	BIOSCharsInt5h BIOSChars = 1 << 26
	// Int 9h, 8042 keyboard services are supported.
	BIOSCharsInt9h BIOSChars = 1 << 27
	// Int 14h, serial services are supported.
	BIOSCharsInt14h BIOSChars = 1 << 28
	// Int 17h, printer services are supported.
	BIOSCharsInt17h BIOSChars = 1 << 29
	// Int 10h, CGA/Mono Video Services are supported.
	BIOSCharsInt10h BIOSChars = 1 << 30
	// NEC PC-98.
	BIOSCharsNECPC98 BIOSChars = 1 << 31
)

var biosCharToString = map[BIOSChars]string{
	BIOSCharsReserved:               "Reserved",
	BIOSCharsReserved2:              "Reserved",
	BIOSCharsUnknown:                "Unknown",
	BIOSCharsAreNotSupported:        "BIOS characteristics not supported",
	BIOSCharsISA:                    "ISA is supported",
	BIOSCharsMCA:                    "MCA is supported",
	BIOSCharsEISA:                   "EISA is supported",
	BIOSCharsPCI:                    "PCI is supported",
	BIOSCharsPCMCIA:                 "PC Card (PCMCIA) is supported",
	BIOSCharsPlugAndPlay:            "PNP is supported",
	BIOSCharsAPM:                    "APM is supported",
	BIOSCharsBIOSUpgradeableFlash:   "BIOS is upgradeable",
	BIOSCharsBIOSShadowingIsAllowed: "BIOS shadowing is allowed",
	BIOSCharsVLVESA:                 "VLB is supported",
	BIOSCharsESCD:                   "ESCD support is available",
	BIOSCharsBootFromCD:             "Boot from CD is supported",
	BIOSCharsSelectableBoot:         "Selectable boot is supported",
	BIOSCharsBIOSROMSocketed:        "BIOS ROM is socketed",
	BIOSCharsBootFromPCMCIA:         "Boot from PC Card (PCMCIA) is supported",
	BIOSCharsEDD:                    "EDD is supported",
	BIOSCharsJapaneseFloppyNEC:      "Japanese floppy for NEC 9800 1.2 MB is supported (int 13h)",
	BIOSCharsJapaneseFloppyToshiba:  "Japanese floppy for Toshiba 1.2 MB is supported (int 13h)",
	BIOSChars360KBFloppy:            "5.25\"/360 kB floppy services are supported (int 13h)",
	BIOSChars12MBFloppy:             "5.25\"/1.2 MB floppy services are supported (int 13h)",
	BIOSChars720KBFloppy:            "3.5\"/720 kB floppy services are supported (int 13h)",
	BIOSChars288MBFloppy:            "3.5\"/2.88 MB floppy services are supported (int 13h)",
	BIOSCharsInt5h:                  "Print screen service is supported (int 5h)",
	BIOSCharsInt9h:                  "8042 keyboard services are supported (int 9h)",
	BIOSCharsInt14h:                 "Serial services are supported (int 14h)",
	BIOSCharsInt17h:                 "Printer services are supported (int 17h)",
	BIOSCharsInt10h:                 "CGA/mono video services are supported (int 10h)",
	BIOSCharsNECPC98:                "NEC PC-98",
}

func (v BIOSChars) String() string {
	var lines []string
	for bit := 0; bit < 32; bit++ {
		if v&(1<<bit) != 0 {
			lines = append(lines, "\t\t"+biosCharToString[1<<bit])
		}
	}
	return strings.Join(lines, "\n")
}

// BIOSCharsExt1 is defined in DSP0134 7.1.2.1.
type BIOSCharsExt1 uint8

// BIOSCharsExt1 is defined in DSP0134 7.1.2.1.
const (
	BIOSCharsExt1ACPI               BIOSCharsExt1 = 1 << 0 // ACPI is supported.
	BIOSCharsExt1USBLegacy          BIOSCharsExt1 = 1 << 1 // USB Legacy is supported.
	BIOSCharsExt1AGP                BIOSCharsExt1 = 1 << 2 // AGP is supported.
	BIOSCharsExt1I2OBoot            BIOSCharsExt1 = 1 << 3 // I2O boot is supported.
	BIOSCharsExt1LS120SuperDiskBoot BIOSCharsExt1 = 1 << 4 // LS-120 SuperDisk boot is supported.
	BIOSCharsExt1ATAPIZIPDriveBoot  BIOSCharsExt1 = 1 << 5 // ATAPI ZIP drive boot is supported.
	BIOSCharsExt11394Boot           BIOSCharsExt1 = 1 << 6 // 1394 boot is supported.
	BIOSCharsExt1SmartBattery       BIOSCharsExt1 = 1 << 7 // Smart battery is supported.
)

var biosCharsExt1Map = map[BIOSCharsExt1]string{
	BIOSCharsExt1ACPI:               "ACPI is supported",
	BIOSCharsExt1USBLegacy:          "USB legacy is supported",
	BIOSCharsExt1AGP:                "AGP is supported",
	BIOSCharsExt1I2OBoot:            "I2O boot is supported",
	BIOSCharsExt1LS120SuperDiskBoot: "LS-120 boot is supported",
	BIOSCharsExt1ATAPIZIPDriveBoot:  "ATAPI Zip drive boot is supported",
	BIOSCharsExt11394Boot:           "IEEE 1394 boot is supported",
	BIOSCharsExt1SmartBattery:       "Smart battery is supported",
}

func (v BIOSCharsExt1) String() string {
	var lines []string
	for bit := 0; bit < 8; bit++ {
		if v&(1<<bit) != 0 {
			lines = append(lines, "\t\t"+biosCharsExt1Map[1<<bit])
		}
	}
	return strings.Join(lines, "\n")
}

// BIOSCharsExt2 is defined in DSP0134 7.1.2.2.
type BIOSCharsExt2 uint8

// BIOSCharsExt1 is defined in DSP0134 7.1.2.2.
const (
	BIOSCharsExt2BIOSBootSpecification       BIOSCharsExt2 = 1 << 0 // BIOS Boot Specification is supported.
	BIOSCharsExt2FnNetworkServiceBoot        BIOSCharsExt2 = 1 << 1 // Function key-initiated network service boot is supported.
	BIOSCharsExt2TargetedContentDistribution BIOSCharsExt2 = 1 << 2 // Enable targeted content distribution.
	BIOSCharsExt2UEFISpecification           BIOSCharsExt2 = 1 << 3 // UEFI Specification is supported.
	BIOSCharsExt2SMBIOSTableDescribesVM      BIOSCharsExt2 = 1 << 4 // SMBIOS table describes a virtual machine. (If this bit is not set, no inference can be made
)

var biosCharsExt2Map = map[BIOSCharsExt2]string{
	BIOSCharsExt2BIOSBootSpecification:       "BIOS boot specification is supported",
	BIOSCharsExt2FnNetworkServiceBoot:        "Function key-initiated network boot is supported",
	BIOSCharsExt2TargetedContentDistribution: "Targeted content distribution is supported",
	BIOSCharsExt2UEFISpecification:           "UEFI is supported",
	BIOSCharsExt2SMBIOSTableDescribesVM:      "System is a virtual machine",
}

func (v BIOSCharsExt2) String() string {
	var lines []string
	for bit := 0; bit < 5; bit++ {
		if v&(1<<bit) != 0 {
			lines = append(lines, "\t\t"+biosCharsExt2Map[1<<bit])
		}
	}
	return strings.Join(lines, "\n")
}
