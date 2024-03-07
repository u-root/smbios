// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dmidecode gives access to decoded SMBIOS/DMI structures.
package dmidecode

import (
	"fmt"

	"github.com/u-root/smbios"
)

const (
	outOfSpec = "<OUT OF SPEC>"
)

// Info contains the SMBIOS information.
type Info struct {
	Entry32 *smbios.Entry32
	Entry64 *smbios.Entry64
	Tables  []*smbios.Table
}

// String returns a summary of the SMBIOS version and number of tables.
func (i *Info) String() string {
	return fmt.Sprintf("SMBIOS %d.%d.%d (%d tables)", i.MajorVersion(), i.MinorVersion(), i.DocRev(), len(i.Tables))
}

// ParseInfo parses SMBIOS information from binary data.
func ParseInfo(entryData, tableData []byte) (*Info, error) {
	info := &Info{}
	var err error
	info.Entry32, info.Entry64, err = smbios.ParseEntry(entryData)
	if err != nil {
		return nil, fmt.Errorf("error parsing entry point structure: %v", err)
	}
	for len(tableData) > 0 {
		t, remainder, err := smbios.ParseTable(tableData)
		if err != nil && err != smbios.ErrEndOfTable {
			return nil, err
		}
		info.Tables = append(info.Tables, t)
		tableData = remainder
	}
	return info, nil
}

// MajorVersion return major version of the SMBIOS spec.
func (i *Info) MajorVersion() uint8 {
	if i.Entry64 != nil {
		return i.Entry64.MajorVersion
	}
	if i.Entry32 != nil {
		return i.Entry32.MajorVersion
	}
	return 0
}

// MinorVersion return minor version of the SMBIOS spec.
func (i *Info) MinorVersion() uint8 {
	if i.Entry64 != nil {
		return i.Entry64.MinorVersion
	}
	if i.Entry32 != nil {
		return i.Entry32.MinorVersion
	}
	return 0
}

// DocRev return document revision of the SMBIOS spec.
func (i *Info) DocRev() uint8 {
	if i.Entry64 != nil {
		return i.Entry64.DocRev
	}
	return 0
}

// GetTablesByType returns tables of specific type.
func (i *Info) GetTablesByType(tt smbios.TableType) []*smbios.Table {
	var res []*smbios.Table
	for _, t := range i.Tables {
		if t.Type == tt {
			res = append(res, t)
		}
	}
	return res
}

// GetBIOSInfo returns the Bios Info (type 0) table, if present.
func (i *Info) GetBIOSInfo() (*BIOSInfo, error) {
	bt := i.GetTablesByType(smbios.TableTypeBIOSInfo)
	if len(bt) == 0 {
		return nil, smbios.ErrTableNotFound
	}
	// There can only be one of these.
	return ParseBIOSInfo(bt[0])
}

// GetSystemInfo returns the System Info (type 1) table, if present.
func (i *Info) GetSystemInfo() (*SystemInfo, error) {
	bt := i.GetTablesByType(smbios.TableTypeSystemInfo)
	if len(bt) == 0 {
		return nil, smbios.ErrTableNotFound
	}
	// There can only be one of these.
	return ParseSystemInfo(bt[0])
}

// GetBaseboardInfo returns all the Baseboard Info (type 2) tables present.
func (i *Info) GetBaseboardInfo() ([]*BaseboardInfo, error) {
	var res []*BaseboardInfo
	for _, t := range i.GetTablesByType(smbios.TableTypeBaseboardInfo) {
		bi, err := ParseBaseboardInfo(t)
		if err != nil {
			return nil, err
		}
		res = append(res, bi)
	}
	return res, nil
}

// GetChassisInfo returns all the Chassis Info (type 3) tables present.
func (i *Info) GetChassisInfo() ([]*ChassisInfo, error) {
	var res []*ChassisInfo
	for _, t := range i.GetTablesByType(smbios.TableTypeChassisInfo) {
		ci, err := ParseChassisInfo(t)
		if err != nil {
			return nil, err
		}
		res = append(res, ci)
	}
	return res, nil
}

// GetProcessorInfo returns all the Processor Info (type 4) tables present.
func (i *Info) GetProcessorInfo() ([]*ProcessorInfo, error) {
	var res []*ProcessorInfo
	for _, t := range i.GetTablesByType(smbios.TableTypeProcessorInfo) {
		pi, err := ParseProcessorInfo(t)
		if err != nil {
			return nil, err
		}
		res = append(res, pi)
	}
	return res, nil
}

// GetCacheInfo returns all the Cache Info (type 7) tables present.
func (i *Info) GetCacheInfo() ([]*CacheInfo, error) {
	var res []*CacheInfo
	for _, t := range i.GetTablesByType(smbios.TableTypeCacheInfo) {
		ci, err := ParseCacheInfo(t)
		if err != nil {
			return nil, err
		}
		res = append(res, ci)
	}
	return res, nil
}

// GetSystemSlots returns all the System Slots (type 9) tables present.
func (i *Info) GetSystemSlots() ([]*SystemSlots, error) {
	var res []*SystemSlots
	for _, t := range i.GetTablesByType(smbios.TableTypeSystemSlots) {
		ss, err := ParseSystemSlots(t)
		if err != nil {
			return nil, err
		}
		res = append(res, ss)
	}
	return res, nil
}

// GetMemoryDevices returns all the Memory Device (type 17) tables present.
func (i *Info) GetMemoryDevices() ([]*MemoryDevice, error) {
	var res []*MemoryDevice
	for _, t := range i.GetTablesByType(smbios.TableTypeMemoryDevice) {
		ci, err := NewMemoryDevice(t)
		if err != nil {
			return nil, err
		}
		res = append(res, ci)
	}
	return res, nil
}

// GetIPMIDeviceInfo returns all the IPMI Device Info (type 38) tables present.
func (i *Info) GetIPMIDeviceInfo() ([]*IPMIDeviceInfo, error) {
	var res []*IPMIDeviceInfo
	for _, t := range i.GetTablesByType(smbios.TableTypeIPMIDeviceInfo) {
		d, err := ParseIPMIDeviceInfo(t)
		if err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, nil
}

// GetTPMDevices returns all the TPM Device (type 43) tables present.
func (i *Info) GetTPMDevices() ([]*TPMDevice, error) {
	var res []*TPMDevice
	for _, t := range i.GetTablesByType(smbios.TableTypeTPMDevice) {
		d, err := NewTPMDevice(t)
		if err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	return res, nil
}

func kmgt(v uint64) string {
	switch {
	case v >= 1024*1024*1024*1024 && v%(1024*1024*1024*1024) == 0:
		return fmt.Sprintf("%d TB", v/(1024*1024*1024*1024))
	case v >= 1024*1024*1024 && v%(1024*1024*1024) == 0:
		return fmt.Sprintf("%d GB", v/(1024*1024*1024))
	case v >= 1024*1024 && v%(1024*1024) == 0:
		return fmt.Sprintf("%d MB", v/(1024*1024))
	case v >= 1024 && v%1024 == 0:
		return fmt.Sprintf("%d kB", v/1024)
	default:
		return fmt.Sprintf("%d bytes", v)
	}
}
