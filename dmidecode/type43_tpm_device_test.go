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

func TestTPMDeviceString(t *testing.T) {
	tests := []struct {
		name string
		val  TPMDevice
		want string
	}{
		{
			name: "Infineon TPM",
			val: TPMDevice{
				VendorID:         TPMDeviceVendorID{0x0, 'X', 'F', 'I'},
				MajorSpecVersion: 1,
				MinorSpecVersion: 7,
				FirmwareVersion1: 2,
				FirmwareVersion2: 3,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristics(8),
				OEMDefined:       2,
			},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Vendor ID: IFX
	Specification Version: 1.7
	Firmware Revision: 0.0
	Description: Test TPM
	Characteristics:
		Family configurable via firmware update
	OEM-specific Info: 0x00000002`,
		},
		{
			name: "Random TPM",
			val: TPMDevice{
				VendorID:         TPMDeviceVendorID{'A', 'B', 'C', 'D'},
				MajorSpecVersion: 2,
				MinorSpecVersion: 9,
				FirmwareVersion1: 1,
				FirmwareVersion2: 9,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristics(16),
				OEMDefined:       2,
			},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Vendor ID: ABCD
	Specification Version: 2.9
	Firmware Revision: 0.1
	Description: Test TPM
	Characteristics:
		Family configurable via platform software support
	OEM-specific Info: 0x00000002`,
		},
		{
			name: "Random TPM #2",
			val: TPMDevice{
				VendorID:         TPMDeviceVendorID{'A', 'B', 'C', 'D'},
				MajorSpecVersion: 2,
				MinorSpecVersion: 9,
				FirmwareVersion1: 1,
				FirmwareVersion2: 9,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristics(32),
				OEMDefined:       2,
			},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Vendor ID: ABCD
	Specification Version: 2.9
	Firmware Revision: 0.1
	Description: Test TPM
	Characteristics:
		Family configurable via OEM proprietary mechanism
	OEM-specific Info: 0x00000002`,
		},
		{
			name: "Random TPM #2",
			val: TPMDevice{
				VendorID:         TPMDeviceVendorID{'A', 'B', 'C', 'D'},
				MajorSpecVersion: 2,
				MinorSpecVersion: 9,
				FirmwareVersion1: 1,
				FirmwareVersion2: 9,
				Description:      "Test TPM",
				Characteristics:  TPMDeviceCharacteristics(4),
				OEMDefined:       2,
			},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Vendor ID: ABCD
	Specification Version: 2.9
	Firmware Revision: 0.1
	Description: Test TPM
	Characteristics:
		TPM Device characteristics not supported
	OEM-specific Info: 0x00000002`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.val.String()

			if result != tt.want {
				t.Errorf("%q failed. Got: %q, Want: %q", tt.name, result, tt.want)
			}
		})
	}

}

func TestNewTPMDevice(t *testing.T) {
	tests := []struct {
		name  string
		table *smbios.Table
		want  *TPMDevice
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
					Type: smbios.TableTypeTPMDevice,
				},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "Parse valid TPMDevice",
			table: &smbios.Table{
				Header: smbios.Header{
					Type:   smbios.TableTypeTPMDevice,
					Length: 32,
				},
				Data: []byte{
					'G', 'O', 'O', 'G',
					2,
					0,
					0x01, 0x02, 0x03, 0x04,
					0x05, 0x02, 0x03, 0x04,
					0x01,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
				},
				Strings: []string{"TPM"},
			},
			want: &TPMDevice{
				Header: smbios.Header{
					Type:   smbios.TableTypeTPMDevice,
					Length: 32,
				},
				VendorID:         [4]byte{'G', 'O', 'O', 'G'},
				MajorSpecVersion: 2,
				MinorSpecVersion: 0,
				FirmwareVersion1: 0x04030201,
				FirmwareVersion2: 0x04030205,
				Description:      "TPM",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTPMDevice(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseTPMDevice = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTPMDevice = %v, want %v", got, tt.want)
			}
		})
	}
}
