// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/u-root/smbios"
)

// ChassisInfo is defined in DSP0134 7.4.
type ChassisInfo struct {
	smbios.Header      `smbios:"-"`
	Manufacturer       string                   // 04h
	Type               ChassisType              // 05h
	Version            string                   // 06h
	SerialNumber       string                   // 07h
	AssetTagNumber     string                   // 08h
	BootupState        ChassisState             // 09h
	PowerSupplyState   ChassisState             // 0Ah
	ThermalState       ChassisState             // 0Bh
	SecurityStatus     ChassisSecurityStatus    // 0Ch
	OEMInfo            uint32                   // 0Dh
	Height             uint8                    // 11h
	NumberOfPowerCords uint8                    // 12h
	ContainedElements  ChassisContainedElements // 13h
	SKUNumber          string                   // 15h + CEC * CERL
}

// Typ implements Table.Typ.
func (ci ChassisInfo) Typ() smbios.TableType {
	return smbios.TableTypeChassisInfo
}

// ChassisContainedElement is defined in DSP0134 7.4.4.
type ChassisContainedElement struct {
	Type ChassisElementType // 00h
	Min  uint8              // 01h
	Max  uint8              // 02h
}

func (cce ChassisContainedElement) String() string {
	return fmt.Sprintf("%s %d-%d", cce.Type, cce.Min, cce.Max)
}

// ParseChassisInfo parses a generic smbios.Table into ChassisInfo.
func ParseChassisInfo(t *smbios.Table) (*ChassisInfo, error) {
	if t.Type != smbios.TableTypeChassisInfo {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedTableType, t.Type)
	}
	if t.Len() < 0x9 {
		return nil, fmt.Errorf("%w: system info table must be at least %d bytes", io.ErrUnexpectedEOF, 9)
	}
	si := &ChassisInfo{Header: t.Header}
	_, err := parseStruct(t, 0 /* off */, false /* complete */, si)
	if err != nil {
		return nil, err
	}
	return si, nil
}

func (si *ChassisInfo) String() string {
	lockStr := "Not Present"
	if si.Type&0x80 != 0 {
		lockStr = "Present"
	}
	lines := []string{
		si.Header.String(),
		fmt.Sprintf("Manufacturer: %s", smbiosStr(si.Manufacturer)),
		fmt.Sprintf("Type: %s", si.Type),
		fmt.Sprintf("Lock: %s", smbiosStr(lockStr)),
		fmt.Sprintf("Version: %s", smbiosStr(si.Version)),
		fmt.Sprintf("Serial Number: %s", smbiosStr(si.SerialNumber)),
		fmt.Sprintf("Asset Tag: %s", si.AssetTagNumber),
	}
	if si.Header.Length >= 9 { // 2.1+
		lines = append(lines,
			fmt.Sprintf("Boot-up State: %s", si.BootupState),
			fmt.Sprintf("Power Supply State: %s", si.PowerSupplyState),
			fmt.Sprintf("Thermal State: %s", si.ThermalState),
			fmt.Sprintf("Security Status: %s", si.SecurityStatus),
		)
	}
	if si.Header.Length >= 0xd { // 2.3+
		heightStr, numPCStr := "Unspecified", "Unspecified"
		if si.Height != 0 {
			heightStr = fmt.Sprintf("%d U", si.Height)
		}
		if si.NumberOfPowerCords != 0 {
			numPCStr = fmt.Sprintf("%d", si.NumberOfPowerCords)
		}
		lines = append(lines,
			fmt.Sprintf("OEM Information: 0x%08X", si.OEMInfo),
			fmt.Sprintf("Height: %s", heightStr),
			fmt.Sprintf("Number Of Power Cords: %s", numPCStr),
			si.ContainedElements.str(),
		)
	}
	if int(si.Header.Length) > 0x15+len(si.ContainedElements)*3 {
		lines = append(lines,
			fmt.Sprintf("SKU Number: %s", smbiosStr(si.SKUNumber)),
		)
	}
	return strings.Join(lines, "\n\t")
}

// ChassisType is defined in DSP0134 7.4.1.
type ChassisType uint8

// ChassisType values are defined in DSP0134 7.4.1.
const (
	ChassisTypeOther               ChassisType = 0x01 // Other
	ChassisTypeUnknown             ChassisType = 0x02 // Unknown
	ChassisTypeDesktop             ChassisType = 0x03 // Desktop
	ChassisTypeLowProfileDesktop   ChassisType = 0x04 // Low Profile Desktop
	ChassisTypePizzaBox            ChassisType = 0x05 // Pizza Box
	ChassisTypeMiniTower           ChassisType = 0x06 // Mini Tower
	ChassisTypeTower               ChassisType = 0x07 // Tower
	ChassisTypePortable            ChassisType = 0x08 // Portable
	ChassisTypeLaptop              ChassisType = 0x09 // Laptop
	ChassisTypeNotebook            ChassisType = 0x0a // Notebook
	ChassisTypeHandHeld            ChassisType = 0x0b // Hand Held
	ChassisTypeDockingStation      ChassisType = 0x0c // Docking Station
	ChassisTypeAllInOne            ChassisType = 0x0d // All in One
	ChassisTypeSubNotebook         ChassisType = 0x0e // Sub Notebook
	ChassisTypeSpacesaving         ChassisType = 0x0f // Space-saving
	ChassisTypeLunchBox            ChassisType = 0x10 // Lunch Box
	ChassisTypeMainServerChassis   ChassisType = 0x11 // Main Server Chassis
	ChassisTypeExpansionChassis    ChassisType = 0x12 // Expansion Chassis
	ChassisTypeSubChassis          ChassisType = 0x13 // SubChassis
	ChassisTypeBusExpansionChassis ChassisType = 0x14 // Bus Expansion Chassis
	ChassisTypePeripheralChassis   ChassisType = 0x15 // Peripheral Chassis
	ChassisTypeRAIDChassis         ChassisType = 0x16 // RAID Chassis
	ChassisTypeRackMountChassis    ChassisType = 0x17 // Rack Mount Chassis
	ChassisTypeSealedcasePC        ChassisType = 0x18 // Sealed-case PC
	ChassisTypeMultisystemChassis  ChassisType = 0x19 // Multi-system chassis
	ChassisTypeCompactPCI          ChassisType = 0x1a // Compact PCI
	ChassisTypeAdvancedTCA         ChassisType = 0x1b // Advanced TCA
	ChassisTypeBlade               ChassisType = 0x1c // Blade
	ChassisTypeBladeChassis        ChassisType = 0x1d // Blade Chassis
	ChassisTypeTablet              ChassisType = 0x1e // Tablet
	ChassisTypeConvertible         ChassisType = 0x1f // Convertible
	ChassisTypeDetachable          ChassisType = 0x20 // Detachable
	ChassisTypeIoTGateway          ChassisType = 0x21 // IoT Gateway
	ChassisTypeEmbeddedPC          ChassisType = 0x22 // Embedded PC
	ChassisTypeMiniPC              ChassisType = 0x23 // Mini PC
	ChassisTypeStickPC             ChassisType = 0x24 // Stick PC
)

var chassisTypeStr = []string{
	"Other",
	"Unknown",
	"Desktop",
	"Low Profile Desktop",
	"Pizza Box",
	"Mini Tower",
	"Tower",
	"Portable",
	"Laptop",
	"Notebook",
	"Hand Held",
	"Docking Station",
	"All In One",
	"Sub Notebook",
	"Space-saving",
	"Lunch Box",
	"Main Server Chassis",
	"Expansion Chassis",
	"Sub Chassis",
	"Bus Expansion Chassis",
	"Peripheral Chassis",
	"RAID Chassis",
	"Rack Mount Chassis",
	"Sealed-case PC",
	"Multi-system",
	"CompactPCI",
	"AdvancedTCA",
	"Blade",
	"Blade Chassis",
	"Tablet",
	"Convertible",
	"Detachable",
	"IoT Gateway",
	"Embedded PC",
	"Mini PC",
	"Stick PC",
}

func (v ChassisType) String() string {
	idx := v&0x7f - 1
	if int(idx) < len(chassisTypeStr) {
		return chassisTypeStr[idx]
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// ChassisState is defined in DSP0134 7.4.2.
type ChassisState uint8

// ChassisState values are defined in DSP0134 7.4.2.
const (
	ChassisStateOther          ChassisState = 0x01 // Other
	ChassisStateUnknown        ChassisState = 0x02 // Unknown
	ChassisStateSafe           ChassisState = 0x03 // Safe
	ChassisStateWarning        ChassisState = 0x04 // Warning
	ChassisStateCritical       ChassisState = 0x05 // Critical
	ChassisStateNonrecoverable ChassisState = 0x06 // Non-recoverable
)

var chassisStateStr = map[ChassisState]string{
	ChassisStateOther:          "Other",
	ChassisStateUnknown:        "Unknown",
	ChassisStateSafe:           "Safe",
	ChassisStateWarning:        "Warning",
	ChassisStateCritical:       "Critical",
	ChassisStateNonrecoverable: "Non-recoverable",
}

func (v ChassisState) String() string {
	if name, ok := chassisStateStr[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// ChassisSecurityStatus is defined in DSP0134 7.4.3.
type ChassisSecurityStatus uint8

// ChassisSecurityStatus values are defined in DSP0134 7.4.3.
const (
	ChassisSecurityStatusOther                      ChassisSecurityStatus = 0x01 // Other
	ChassisSecurityStatusUnknown                    ChassisSecurityStatus = 0x02 // Unknown
	ChassisSecurityStatusNone                       ChassisSecurityStatus = 0x03 // None
	ChassisSecurityStatusExternalInterfaceLockedOut ChassisSecurityStatus = 0x04 // External interface locked out
	ChassisSecurityStatusExternalInterfaceEnabled   ChassisSecurityStatus = 0x05 // External interface enabled
)

var chassisSecurityStatusStr = map[ChassisSecurityStatus]string{
	ChassisSecurityStatusOther:                      "Other",
	ChassisSecurityStatusUnknown:                    "Unknown",
	ChassisSecurityStatusNone:                       "None",
	ChassisSecurityStatusExternalInterfaceLockedOut: "External Interface Locked Out",
	ChassisSecurityStatusExternalInterfaceEnabled:   "External Interface Enabled",
}

func (v ChassisSecurityStatus) String() string {
	if name, ok := chassisSecurityStatusStr[v]; ok {
		return name
	}
	return fmt.Sprintf("%#x", uint8(v))
}

// ChassisElementType is defined in DSP0134 7.4.4.
type ChassisElementType uint8

func (v ChassisElementType) String() string {
	if v&0x80 != 0 {
		return smbios.TableType(v & 0x7f).String()
	}
	return BoardType(v & 0x7f).String()
}

// ChassisContainedElements are defined by DSP0134 7.4.4.
type ChassisContainedElements []ChassisContainedElement

func (cec ChassisContainedElements) String() string {
	lines := []string{fmt.Sprintf("Contained Elements: %d", len(cec))}
	for _, e := range cec {
		lines = append(lines, fmt.Sprintf("\t%s", e))
	}
	return strings.Join(lines, "\n")
}

func (cec ChassisContainedElements) str() string {
	lines := []string{fmt.Sprintf("Contained Elements: %d", len(cec))}
	for _, e := range cec {
		lines = append(lines, fmt.Sprintf("\t\t%s", e))
	}
	return strings.Join(lines, "\n")
}

// WriteField writes contained elements as defined by DSP0134 Section 7.4.4.
func (cec *ChassisContainedElements) WriteField(t *smbios.Table) (int, error) {
	num := len(*cec)
	if num > math.MaxUint8 {
		return 0, fmt.Errorf("%w: too many object handles defined, can be maximum of 256", ErrInvalidArg)
	}
	t.WriteByte(uint8(num))

	if num == 0 {
		t.WriteByte(0)
		return 2, nil
	}
	t.WriteByte(3)
	for _, el := range *cec {
		if err := binary.Write(t, binary.LittleEndian, el); err != nil {
			return 0, err
		}
	}
	return 2 + 3*num, nil
}

// ParseField parses contained elements as defined by DSP0134 Section 7.4.4.
func (cec *ChassisContainedElements) ParseField(t *smbios.Table, off int) (int, error) {
	num, err := t.GetByteAt(off)
	if err != nil {
		return off, err
	}
	off++

	size, err := t.GetByteAt(off)
	if err != nil {
		return off, err
	}
	off++

	if num == 0 && size == 0 {
		return off, nil
	}
	if size != 3 {
		return off, fmt.Errorf("%w: unexpected chassis contained element size %d, support 3", os.ErrInvalid, size)
	}
	for i := uint8(0); i < num; i++ {
		var e ChassisContainedElement
		if err := binary.Read(io.NewSectionReader(t, int64(off), 3), binary.LittleEndian, &e); err != nil {
			return off, err
		}
		*cec = append(*cec, e)
		off += 3
	}
	return off, nil
}
