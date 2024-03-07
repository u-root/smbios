// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

// TableType specifies the DMI type of the table.
// Types are defined in DMTF DSP0134.
type TableType uint8

// Supported table types.
const (
	TableTypeBIOSInfo       TableType = 0
	TableTypeSystemInfo     TableType = 1
	TableTypeBaseboardInfo  TableType = 2
	TableTypeChassisInfo    TableType = 3
	TableTypeProcessorInfo  TableType = 4
	TableTypeCacheInfo      TableType = 7
	TableTypeSystemSlots    TableType = 9
	TableTypeMemoryDevice   TableType = 17
	TableTypeIPMIDeviceInfo TableType = 38
	TableTypeTPMDevice      TableType = 43
	TableTypeInactive       TableType = 126
	TableTypeEndOfTable     TableType = 127
)

func (t TableType) String() string {
	switch t {
	case TableTypeBIOSInfo:
		return "BIOS Information"
	case TableTypeSystemInfo:
		return "System Information"
	case TableTypeBaseboardInfo:
		return "Base Board Information"
	case TableTypeChassisInfo:
		return "Chassis Information"
	case TableTypeProcessorInfo:
		return "Processor Information"
	case TableTypeCacheInfo:
		return "Cache Information"
	case TableTypeSystemSlots:
		return "System Slots"
	case TableTypeMemoryDevice:
		return "Memory Device"
	case TableTypeIPMIDeviceInfo:
		return "IPMI Device Information"
	case TableTypeTPMDevice:
		return "TPM Device"
	case TableTypeInactive:
		return "Inactive"
	case TableTypeEndOfTable:
		return "End Of Table"
	default:
		if t >= 0x80 {
			return "OEM-specific Type"
		}
		return "Unsupported"
	}
}
