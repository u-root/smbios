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

func TestGetFamily(t *testing.T) {
	tests := []struct {
		name string
		val  ProcessorInfo
		want uint16
	}{
		{
			name: "Type not 0xfe",
			val: ProcessorInfo{
				Family:  0xff,
				Family2: 0xaa,
			},
			want: 0xff,
		},
		{
			name: "Type 0xfe wrong length",
			val: ProcessorInfo{
				Family:  0xfe,
				Family2: 0xaa,
			},
			want: 0xfe,
		},
		{
			name: "Type 0xfe",
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3a,
				},
				Family:  0xfe,
				Family2: 0xaa,
			},
			want: 0xaa,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.val.GetFamily()
			if result != ProcessorFamily(tt.want) {
				t.Errorf("GeFamily(): '%v', want '%v'", result, tt.want)
			}
		})
	}
}

func TestGetVoltage(t *testing.T) {
	tests := []struct {
		name string
		val  ProcessorInfo
		want float32
	}{
		{
			name: "5V",
			val: ProcessorInfo{
				Voltage: 0x1,
			},
			want: 5,
		},
		{
			name: "3.3V",
			val: ProcessorInfo{
				Voltage: 0x2,
			},
			want: 3.3,
		},
		{
			name: "2.9V",
			val: ProcessorInfo{
				Voltage: 0x4,
			},
			want: 2.9,
		},
		{
			name: "Unknown",
			val: ProcessorInfo{
				Voltage: 0xFF,
			},
			want: 12.7,
		},
		{
			name: "Zero",
			val: ProcessorInfo{
				Voltage: 0x10,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.val.GetVoltage()

			if result != tt.want {
				t.Errorf("GetVoltage(): '%v', want '%v'", result, tt.want)
			}
		})
	}
}

func TestGetCoreCount(t *testing.T) {
	for _, tt := range []struct {
		val  ProcessorInfo
		want int
	}{
		{
			val: ProcessorInfo{
				CoreCount:  4,
				CoreCount2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				CoreCount:  4,
				CoreCount2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				CoreCount:  0xff,
				CoreCount2: 8,
			},
			want: 8,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := tt.val.GetCoreCount()
			if got != tt.want {
				t.Errorf("GetCoreCount = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCoreEnabled(t *testing.T) {
	for _, tt := range []struct {
		val  ProcessorInfo
		want int
	}{
		{
			val: ProcessorInfo{
				CoreEnabled:  4,
				CoreEnabled2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				CoreEnabled:  4,
				CoreEnabled2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				CoreEnabled:  0xff,
				CoreEnabled2: 0x8ff,
			},
			want: 0x8ff,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := tt.val.GetCoreEnabled()
			if got != tt.want {
				t.Errorf("GetCoreEnabled = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetThreadCount(t *testing.T) {
	for _, tt := range []struct {
		val  ProcessorInfo
		want int
	}{
		{
			val: ProcessorInfo{
				ThreadCount:  4,
				ThreadCount2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				ThreadCount:  4,
				ThreadCount2: 8,
			},
			want: 4,
		},
		{
			val: ProcessorInfo{
				Header: smbios.Header{
					Length: 0x3c,
				},
				ThreadCount:  0xff,
				ThreadCount2: 0x8ff,
			},
			want: 0x8ff,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := tt.val.GetThreadCount()
			if got != tt.want {
				t.Errorf("GetThreadCount = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessorTypeString(t *testing.T) {
	for _, tt := range []struct {
		val  ProcessorType
		want string
	}{
		{
			val:  ProcessorType(1),
			want: "Other",
		},
		{
			val:  ProcessorType(6),
			want: "Video Processor",
		},
		{
			val:  ProcessorType(8),
			want: "0x8",
		},
	} {
		t.Run("", func(t *testing.T) {
			got := tt.val.String()
			if got != tt.want {
				t.Errorf("ProcessorType = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestProcessorUpgradeString(t *testing.T) {
	resultStrings := []string{
		"Other",
		"Unknown",
		"Daughter Board",
		"ZIF Socket",
		"Replaceable Piggy Back",
		"None",
		"LIF Socket",
		"Slot 1",
		"Slot 2",
		"370-pin Socket",
		"Slot A",
		"Slot M",
		"Socket 423",
		"Socket A (Socket 462)",
		"Socket 478",
		"Socket 754",
		"Socket 940",
		"Socket 939",
		"Socket mPGA604",
		"Socket LGA771",
		"Socket LGA775",
		"Socket S1",
		"Socket AM2",
		"Socket F (1207)",
		"Socket LGA1366",
		"Socket G34",
		"Socket AM3",
		"Socket C32",
		"Socket LGA1156",
		"Socket LGA1567",
		"Socket PGA988A",
		"Socket BGA1288",
		"Socket rPGA988B",
		"Socket BGA1023",
		"Socket BGA1224",
		"Socket BGA1155",
		"Socket LGA1356",
		"Socket LGA2011",
		"Socket FS1",
		"Socket FS2",
		"Socket FM1",
		"Socket FM2",
		"Socket LGA2011-3",
		"Socket LGA1356-3",
		"Socket LGA1150",
		"Socket BGA1168",
		"Socket BGA1234",
		"Socket BGA1364",
		"Socket AM4",
		"Socket LGA1151",
		"Socket BGA1356",
		"Socket BGA1440",
		"Socket BGA1515",
		"Socket LGA3647-1",
		"Socket SP3",
		"Socket SP3r2",
		"Socket LGA2066",
		"Socket BGA1392",
		"Socket BGA1510",
		"Socket BGA1528",
	}

	for id := range resultStrings {
		ProcessorUpgr := ProcessorUpgrade(id + 1)
		if ProcessorUpgr.String() != resultStrings[id] {
			t.Errorf("ProcessorUpgrade().String(): '%s', want '%s'", ProcessorUpgr.String(), resultStrings[id])
		}
	}
}

func TestProcessorCharacteristicsString(t *testing.T) {
	tests := []struct {
		name string
		val  ProcessorCharacteristics
		want string
	}{
		{
			name: "Reserved Characteristics",
			val:  ProcessorCharacteristics(0x01),
			want: "		Reserved",
		},
		{
			name: "Unknown Characteristics",
			val:  ProcessorCharacteristics(0x02),
			want: "		Unknown",
		},
		{
			name: "64-bit capable Characteristics",
			val:  ProcessorCharacteristics(0x04),
			want: "		64-bit capable",
		},
		{
			name: "Multi-Core Characteristics",
			val:  ProcessorCharacteristics(0x08),
			want: "		Multi-Core",
		},
		{
			name: "Hardware Thread Characteristics",
			val:  ProcessorCharacteristics(0x10),
			want: "		Hardware Thread",
		},
		{
			name: "Execute Protection Characteristics",
			val:  ProcessorCharacteristics(0x20),
			want: "		Execute Protection",
		},
		{
			name: "Enhanced Virtualization Characteristics",
			val:  ProcessorCharacteristics(0x40),
			want: "		Enhanced Virtualization",
		},
		{
			name: "Power/Performance Control Characteristics",
			val:  ProcessorCharacteristics(0x80),
			want: "		Power/Performance Control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.String() != tt.want {
				t.Errorf("ProcessorCharacteristics().String(): '%s', want '%s'", tt.val.String(), tt.want)
			}
		})
	}
}

func TestParseProcessorInfo(t *testing.T) {
	tests := []struct {
		name  string
		table *smbios.Table
		want  *ProcessorInfo
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
					Type: smbios.TableTypeProcessorInfo,
				},
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "Parse valid ProcessorInfo",
			table: &smbios.Table{
				Header: smbios.Header{
					Type: smbios.TableTypeProcessorInfo,
				},
				Data: []byte{
					0x00, // SocketDesig
					0x01, // type
					0x03, // family
					0x00, // manufacturer
					0x01, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, // id
					0x00,       // version
					0x01,       // voltage
					0x01, 0x00, // clock
					0x02, 0x00, // max speed
					0x03, 0x00, // current speed
					0x0a, // status
					0x0f, // upgrade
				},
			},
			want: &ProcessorInfo{
				Header: smbios.Header{
					Type: smbios.TableTypeProcessorInfo,
				},
				Type:          0x1,
				Family:        0x3,
				ID:            0x1,
				Voltage:       0x1,
				ExternalClock: 0x1,
				MaxSpeed:      0x2,
				CurrentSpeed:  0x3,
				Status:        0xa,
				Upgrade:       0xf,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProcessorInfo(tt.table)
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseProcessorInfo = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProcessorInfo = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessorInfoString(t *testing.T) {
	for _, tt := range []struct {
		val  ProcessorInfo
		want string
	}{
		{
			val: ProcessorInfo{},
			want: `Handle 0x0000, DMI type 0, 0 bytes
BIOS Information
	Socket Designation: Not Specified
	Type: 0x0
	Family: 0x0
	Manufacturer: Not Specified
	ID: 00 00 00 00 00 00 00 00
	Version: Not Specified
	Voltage: 0.0 V
	External Clock: Unknown
	Max Speed: Unknown
	Current Speed: Unknown
	Status: Unpopulated
	Upgrade: 0x0`,
		},
		{
			val: ProcessorInfo{
				Type:          ProcessorTypeCentralProcessor,
				Family:        0x10,
				Manufacturer:  "Google",
				ID:            0x1234567812345678,
				L1CacheHandle: 0x1337,
				L2CacheHandle: 0xDEAD,
				L3CacheHandle: 0xBEEF,
				Header: smbios.Header{
					Length: 0x3c,
					Type:   smbios.TableTypeProcessorInfo,
				},
				ThreadCount: 8,
			},
			want: `Handle 0x0000, DMI type 4, 60 bytes
Processor Information
	Socket Designation: Not Specified
	Type: Central Processor
	Family: Pentium II Xeon
	Manufacturer: Google
	ID: 78 56 34 12 78 56 34 12
	Signature: Type 1, Family 41, Model 71, Stepping 8
	Flags:
		PSE (Page size extension)
		TSC (Time stamp counter)
		MSR (Model specific registers)
		PAE (Physical address extension)
		APIC (On-chip APIC hardware supported)
		MTRR (Memory type range registers)
		MCA (Machine check architecture)
		PSN (Processor serial number present and enabled)
		DS (Debug store)
		SSE (Streaming SIMD extensions)
		HTT (Multi-threading)
	Version: Not Specified
	Voltage: 0.0 V
	External Clock: Unknown
	Max Speed: Unknown
	Current Speed: Unknown
	Status: Unpopulated
	Upgrade: 0x0
	L1 Cache Handle: 0x1337
	L2 Cache Handle: 0xDEAD
	L3 Cache Handle: 0xBEEF
	Serial Number: Not Specified
	Asset Tag: Not Specified
	Part Number: Not Specified
	Core Count: 0
	Core Enabled: 0
	Thread Count: 8
	Characteristics:
		`,
		},
	} {
		t.Run("", func(t *testing.T) {
			if tt.val.String() != tt.want {
				t.Errorf("ProcessorInfo = %s, want %s", tt.val.String(), tt.want)
			}
		})
	}
}
