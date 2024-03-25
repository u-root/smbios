// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"errors"
	"fmt"

	"github.com/u-root/smbios"
)

// ErrUnsupportedTableType is returned by ParseTypedTable if this table type is not supported and cannot be parsed.
var ErrUnsupportedTableType = errors.New("unsupported table type")

// ParseTypedTable parses generic Table into a typed struct.
func ParseTypedTable(t *smbios.Table) (fmt.Stringer, error) {
	switch t.Type {
	case smbios.TableTypeBIOSInfo: // 0
		return ParseBIOSInfo(t)
	case smbios.TableTypeSystemInfo: // 1
		return ParseSystemInfo(t)
	case smbios.TableTypeBaseboardInfo: // 2
		return ParseBaseboardInfo(t)
	case smbios.TableTypeChassisInfo: // 3
		return ParseChassisInfo(t)
	case smbios.TableTypeProcessorInfo: // 4
		return ParseProcessorInfo(t)
	case smbios.TableTypeCacheInfo: // 7
		return ParseCacheInfo(t)
	case smbios.TableTypeSystemSlots: // 9
		return ParseSystemSlots(t)
	case smbios.TableTypeMemoryDevice: // 17
		return ParseMemoryDevice(t)
	case smbios.TableTypeIPMIDeviceInfo: // 38
		return ParseIPMIDeviceInfo(t)
	case smbios.TableTypeTPMDevice: // 43
		return NewTPMDevice(t)
	case smbios.TableTypeInactive: // 126
		return NewInactiveTable(t)
	case smbios.TableTypeEndOfTable: // 127
		return NewEndOfTable(t)
	}
	return nil, ErrUnsupportedTableType
}
