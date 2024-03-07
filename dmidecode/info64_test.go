// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/u-root/smbios"
)

func Test64ParseInfoHeaderMalformed(t *testing.T) {
	data, err := os.ReadFile("./testdata/smbios_table.bin")
	if err != nil {
		t.Errorf("error reading mockup smbios tables: %v", err)
	}

	entryData := data[:10]
	data = data[32:]

	_, err = ParseInfo(entryData, data)
	if err == nil {
		t.Errorf("error parsing info data: %v", err)
	}
}

func Test64MajorVersion(t *testing.T) {
	info, err := setupMockData()
	if err != nil {
		t.Errorf("error parsing info data: %v", err)
	}
	major, minor, rev := info.Entry.Version()
	if major != 3 {
		t.Errorf("major version = %d, want 3", major)
	}
	if minor != 1 {
		t.Errorf("minor version = %d, want 1", minor)
	}
	if rev != 1 {
		t.Errorf("doc revision = %d, want 1", rev)
	}
}

func Test64GetTablesByType(t *testing.T) {
	info, err := setupMockData()
	if err != nil {
		t.Errorf("error parsing info data: %v", err)
	}

	table := info.GetTablesByType(smbios.TableTypeBIOSInfo)
	if table == nil {
		t.Errorf("unable to get type")
	}
	if table != nil {
		if table[0].Header.Type != smbios.TableTypeBIOSInfo {
			t.Errorf("Wrong type. Got %v but want %v", smbios.TableTypeBIOSInfo, table[0].Header.Type)
		}
	}
}

func setupMockData() (*Info, error) {
	data, err := os.ReadFile("./testdata/smbios_table.bin")
	if err != nil {
		return nil, err
	}

	entryData := data[:32]
	data = data[32:]

	info, err := ParseInfo(entryData, data)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func FuzzParseInfo(f *testing.F) {
	seeds, err := filepath.Glob("testdata/*.bin")
	if err != nil {
		f.Fatalf("failed to find seed corpora files: %v", err)
	}

	for _, seed := range seeds {
		seedBytes, err := os.ReadFile(seed)
		if err != nil {
			f.Fatalf("failed read seed corpora from files %v: %v", seed, err)
		}

		f.Add(seedBytes)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 64 || len(data) > 4096 {
			return
		}

		entryData := data[:32]
		data = data[32:]

		info, err := ParseInfo(entryData, data)
		if err != nil {
			return
		}
		entry, err := info.Entry.MarshalBinary()
		if err != nil {
			t.Fatalf("failed to unmarshal entry data")
		}

		reparsedInfo, err := ParseInfo(entry, data)
		if err != nil {
			t.Fatalf("failed to reparse the SMBIOS info struct")
		}
		if !reflect.DeepEqual(info, reparsedInfo) {
			t.Errorf("expected: %#v\ngot:%#v", info, reparsedInfo)
		}
	})
}

func Test64Len(t *testing.T) {
	info, err := setupMockData()
	if err != nil {
		t.Errorf("error parsing info Data: %v", err)
	}

	if info.Tables != nil {
		if info.Tables[0].Len() != 14 {
			t.Errorf("Wrong length: Got %d want %d", info.Tables[0].Len(), 14)
		}
	}
}

func Test64String(t *testing.T) {

	tableString := `Handle 0x0000, DMI type 222, 14 bytes
OEM-specific Type
	Header and Data:
		DE 0E 00 00 01 99 00 03 10 01 20 02 30 03
	Strings:
		Memory Init Complete
		End of DXE Phase
		BIOS Boot Complete`

	info, err := setupMockData()

	if err != nil {
		t.Errorf("error parsing info Data: %v", err)
	}

	if info.Tables != nil {
		if info.Tables[0].String() != tableString {
			t.Errorf("Wrong length: Got %s want %s", info.Tables[0].String(), tableString)
		}
	}
}

func TestKmgt(t *testing.T) {

	tests := []struct {
		name   string
		value  uint64
		expect string
	}{
		{
			name:   "Just bytes",
			value:  512,
			expect: "512 bytes",
		},
		{
			name:   "Two Kb",
			value:  2 * 1024,
			expect: "2 kB",
		},
		{
			name:   "512 MB",
			value:  512 * 1024 * 1024,
			expect: "512 MB",
		},
		{
			name:   "8 GB",
			value:  8 * 1024 * 1024 * 1024,
			expect: "8 GB",
		},
		{
			name:   "3 TB",
			value:  3 * 1024 * 1024 * 1024 * 1024,
			expect: "3 TB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if kmgt(tt.value) != tt.expect {
				t.Errorf("kgmt(): %v - want '%v'", kmgt(tt.value), tt.expect)
			}
		})
	}
}
