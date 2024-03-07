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

var tableTypeToString = map[TableType]string{
	TableTypeBIOSInfo:       "BIOS Information",
	TableTypeSystemInfo:     "System Information",
	TableTypeBaseboardInfo:  "Base Board Information",
	TableTypeChassisInfo:    "Chassis Information",
	TableTypeProcessorInfo:  "Processor Information",
	TableTypeCacheInfo:      "Cache Information",
	TableTypeSystemSlots:    "System Slots",
	TableTypeMemoryDevice:   "Memory Device",
	TableTypeIPMIDeviceInfo: "IPMI Device Information",
	TableTypeTPMDevice:      "TPM Device",
	TableTypeInactive:       "Inactive",
	TableTypeEndOfTable:     "End Of Table",
}

func (t TableType) String() string {
	if s, ok := tableTypeToString[t]; ok {
		return s
	}
	if t >= 0x80 {
		return "OEM-specific Type"
	}
	return "Unsupported"
}
