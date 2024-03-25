// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
)

var (
	testbinary = "testdata/satellite_pro_l70_testdata.bin"
)

func TestParseSMBIOS(t *testing.T) {
	f, err := os.Open(testbinary)
	if err != nil {
		t.Error(err)
	}
	if _, err := ParseTables(f); err != nil {
		t.Error(err)
	}
}

func TestParseTable(t *testing.T) {
	for _, tt := range []struct {
		b    []byte
		want *Table
		err  error
	}{
		{
			b: []byte{
				0x01,       // type
				0x01,       // data length
				0x00, 0x00, // handle
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x10,       // data length
				0x00, 0x00, // handle
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				'a', // short malformed string
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				'a', 'b', // malformed string
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				0x00, 0x00, // end of strings
			},
			want: &Table{
				Header: Header{
					Type:   1,
					Length: 4,
					Handle: 0,
				},
			},
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				'a', 0x00, // string
				// missing end of strings
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				'a', 0x00, // string
				0x00, // end of strings
			},
			want: &Table{
				Header: Header{
					Type:   1,
					Length: 4,
					Handle: 0,
				},
				Strings: []string{"a"},
			},
		},
		{
			b: []byte{
				0x01,       // type
				0x0a,       // data length
				0x00, 0x00, // handle
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, // formatted data
				'a', 0x00, // string
				'b', 'a', 0x00,
				0x00, // ambiguous
				0x00,
			},
			want: &Table{
				Header: Header{
					Type:   1,
					Length: 10,
					Handle: 0,
				},
				Data: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
				},
				Strings: []string{
					"a",
					"ba",
				},
			},
		},
		{
			b:    nil,
			want: nil,
			err:  io.ErrUnexpectedEOF,
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := ParseTable(bytes.NewReader(tt.b))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTable =\n%v, want\n%v", got, tt.want)
			}
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseTable = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestParseTables(t *testing.T) {
	for _, tt := range []struct {
		b    []byte
		want Tables
		err  error
	}{
		{
			b: []byte{
				0x01,       // type
				0x01,       // data length
				0x00, 0x00, // handle
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				0x00, 0x00, // end of strings
			},
			want: Tables{
				&Table{
					Header: Header{
						Type:   1,
						Length: 4,
						Handle: 0,
					},
				},
			},
		},
		{
			b: []byte{
				0x01,       // type
				0x0a,       // data length
				0x00, 0x00, // handle
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, // formatted data
				'a', 0x00, // string
				'b', 'a', 0x00,
				0x00, // ambiguous
				0x00,
			},
			err: io.ErrUnexpectedEOF,
		},
		{
			b:    nil,
			want: nil,
			err:  nil,
		},
		{
			b: []byte{
				0x01,       // type
				0x04,       // data length
				0x00, 0x00, // handle
				0x00, 0x00, // end of strings

				0x01,       // type
				0x0a,       // data length
				0x00, 0x00, // handle
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, // formatted data
				'a', 0x00, // string
				'b', 'a', 0x00,
				0x00, // end of strings
			},
			want: Tables{
				&Table{
					Header: Header{
						Type:   1,
						Length: 4,
						Handle: 0,
					},
				},
				&Table{
					Header: Header{
						Type:   1,
						Length: 10,
						Handle: 0,
					},
					Data: []byte{1, 2, 3, 4, 5, 6},
					Strings: []string{
						"a",
						"ba",
					},
				},
			},
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := ParseTables(bytes.NewReader(tt.b))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTables =\n%v, want\n%v", got, tt.want)
			}
			if !errors.Is(err, tt.err) {
				t.Errorf("ParseTables = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestGetByteAt(t *testing.T) {
	testStruct := Table{
		Data: []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	for _, tt := range []struct {
		name   string
		offset int
		want   uint8
		err    error
	}{
		{
			name:   "GetByteAt",
			offset: 0,
			want:   1,
		},
		{
			name:   "GetByteAt Wrong Offset",
			offset: 213,
			err:    io.ErrUnexpectedEOF,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStruct.GetByteAt(tt.offset)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetByteAt = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("GetByteAt = %x, want %x", got, tt.want)
			}
		})
	}
}

func TestGetBytesAt(t *testing.T) {
	testStruct := Table{
		Data: []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	for _, tt := range []struct {
		name   string
		offset int
		length int
		want   []byte
		err    error
	}{
		{
			name:   "Get two bytes",
			offset: 0,
			length: 2,
			want:   []byte{1, 0},
		},
		{
			name:   "Wrong Offset",
			offset: 213,
			err:    io.ErrUnexpectedEOF,
		},
		{
			name:   "Read out-of-bounds",
			offset: 7,
			length: 16,
			err:    io.ErrUnexpectedEOF,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStruct.GetBytesAt(tt.offset, tt.length)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetBytesAt = %v, want %v", err, tt.err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("GetBytesAt = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWordAt(t *testing.T) {
	testStruct := Table{
		Data: []byte{
			213, 0, 0, 11,
			12, 0, 0, 0,
			0, 0, 0, 0,
		},
	}
	for _, tt := range []struct {
		name   string
		offset int
		want   uint16
		err    error
	}{
		{
			name:   "Get two bytes",
			offset: 0,
			want:   213,
			err:    nil,
		},
		{
			name:   "edge case offset",
			offset: 11,
			err:    io.ErrUnexpectedEOF,
		},
		{
			name:   "Read position 7",
			offset: 3,
			want:   0xc0b,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStruct.GetWordAt(tt.offset)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetWordAt = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("GetWordAt = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDWordAt(t *testing.T) {
	testStruct := Table{
		Data: []byte{
			1, 0, 0, 0,
			213, 0, 0, 11,
			12, 13, 14, 0,
			0, 0, 0, 0,
		},
	}
	for _, tt := range []struct {
		name   string
		offset int
		want   uint32
		err    error
	}{
		{
			name:   "Get two bytes",
			offset: 0,
			want:   1,
			err:    nil,
		},
		{
			name:   "edge case offset",
			offset: 13,
			err:    io.ErrUnexpectedEOF,
		},
		{
			name:   "Read position 7",
			offset: 7,
			want:   0xe0d0c0b,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStruct.GetDWordAt(tt.offset)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetDWordAt = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("GetDWordAt = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetQWordAt(t *testing.T) {
	testStruct := Table{
		Data: []byte{
			1, 0, 0, 0,
			213, 0, 0, 11,
			12, 13, 14, 15,
			16, 17, 18, 0,
		},
	}
	for _, tt := range []struct {
		name   string
		offset int
		want   uint64
		err    error
	}{
		{
			name:   "Get two bytes",
			offset: 0,
			want:   0xb0000d500000001,
			err:    nil,
		},
		{
			name:   "edge case offset",
			offset: 9,
			want:   0,
			err:    io.ErrUnexpectedEOF,
		},
		{
			name:   "Read position 7",
			offset: 7,
			want:   0x1211100f0e0d0c0b,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStruct.GetQWordAt(tt.offset)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetQWordAt = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("GetQWordAt = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringAt(t *testing.T) {
	testStruct := Table{
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}
	for _, tt := range []struct {
		name   string
		offset int
		want   string
		err    error
	}{
		{
			name:   "Valid offset",
			offset: 0,
			want:   "BIOS Boot Complete",
		},
		{
			name:   "Not Specified",
			offset: 2,
			want:   "",
		},
		{
			name:   "Bad Index",
			offset: 4,
			want:   "<BAD INDEX>",
			err:    io.ErrUnexpectedEOF,
		},
		{
			name:   "out of bounds",
			offset: 20,
			err:    io.ErrUnexpectedEOF,
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := testStruct.GetStringAt(tt.offset)
			if got != tt.want {
				t.Errorf("GetStringAt = %s, want %s", got, tt.want)
			}
			if !errors.Is(err, tt.err) {
				t.Errorf("GetStringAt = %s, want %s", err, tt.err)
			}
		})
	}
}

func TestByType(t *testing.T) {
	tt := Tables{
		&Table{
			Header: Header{
				Type:   1,
				Length: 4,
				Handle: 0,
			},
		},
		&Table{
			Header: Header{
				Type:   1,
				Length: 10,
				Handle: 0,
			},
			Data: []byte{1, 2, 3, 4, 5, 6},
			Strings: []string{
				"a",
				"ba",
			},
		},
	}

	if got := tt.TablesByType(1); !reflect.DeepEqual(got, tt) {
		t.Errorf("ByType(1) = %v, want %v", got, tt)
	}
	if got := tt.TablesByType(2); got != nil {
		t.Errorf("ByType(2) = %v, want %v", got, nil)
	}
	if got := tt.TableByType(2); got != nil {
		t.Errorf("ByType(2) = %v, want %v", got, nil)
	}
	if got := tt.TableByType(1); !reflect.DeepEqual(got, tt[0]) {
		t.Errorf("ByType(1) = %v, want %v", got, tt[0])
	}

	// Test nil safety.
	tt = nil
	if got := tt.TablesByType(2); got != nil {
		t.Errorf("ByType(2) = %v, want %v", got, nil)
	}
	if got := tt.TableByType(2); got != nil {
		t.Errorf("ByType(2) = %v, want %v", got, nil)
	}
}

func TestTableStringLen(t *testing.T) {
	want := `Handle 0x0000, DMI type 222, 14 bytes
OEM-specific Type
	Header and Data:
		DE 0E 00 00 01 99 00 03 10 01 20 02 30 03
	Strings:
		Memory Init Complete
		End of DXE Phase
		BIOS Boot Complete`

	table := &Table{
		Header: Header{
			Type:   222,
			Length: 14,
			Handle: 0,
		},
		Data: []byte{0x01, 0x99, 0x00, 0x03, 0x10, 0x01, 0x20, 0x02, 0x30, 0x03},
		Strings: []string{
			"Memory Init Complete",
			"End of DXE Phase",
			"BIOS Boot Complete",
		},
	}

	if got := table.String(); got != want {
		t.Errorf("Wrong string: Got %s want %s", got, want)
	}
	if got := table.Len(); got != 14 {
		t.Errorf("Wrong length: Got %d want %d", got, 14)
	}
}

func TestTableMarshalBinary(t *testing.T) {
	for _, tt := range []struct {
		name  string
		table Table
		want  []byte
	}{
		{
			name: "full table",
			table: Table{
				Header: Header{
					Type:   224,
					Length: 14,
					Handle: 0,
				},
				Data:    []byte{1, 153, 0, 3, 16, 1, 32, 2, 48, 3},
				Strings: []string{"Memory Init Complete", "End of DXE Phase", "BIOS Boot Complete"},
			},
			want: []byte{
				// Header and Data
				224, 14, 0, 0,
				1, 153, 0, 3, 16, 1, 32, 2, 48, 3,
				// Strings
				77, 101, 109, 111, 114, 121, 32, 73, 110, 105, 116, 32, 67, 111, 109, 112, 108, 101, 116, 101, 0, // Memory Init Complete
				69, 110, 100, 32, 111, 102, 32, 68, 88, 69, 32, 80, 104, 97, 115, 101, 0, // End of DXE Phase
				66, 73, 79, 83, 32, 66, 111, 111, 116, 32, 67, 111, 109, 112, 108, 101, 116, 101, 0, //  BIOS Boot Complete
				0, // Table terminator
			},
		},
		{
			name: "NoString",
			table: Table{
				Header: Header{
					Type:   224,
					Length: 14,
					Handle: 0,
				},
				Data: []byte{1, 153, 0, 3, 16, 1, 32, 2, 48, 3},
			},
			want: []byte{
				// Header and Data
				224, 14, 0, 0,
				1, 153, 0, 3, 16, 1, 32, 2, 48, 3,
				// Strings
				0, // String terminator
				0, // Table terminator
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.table.MarshalBinary()
			if err != nil {
				t.Errorf("MarshalBinary returned error: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Wrong raw data: Got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTableWrite(t *testing.T) {
	var tt Table
	tt.WriteByte(1)
	tt.WriteWord(0x0203)
	tt.WriteDWord(0x04050607)
	tt.WriteQWord(0x0102030405060708)
	tt.WriteBytes([]byte{0, 1})
	tt.WriteString("foo")
	tt.WriteString("")
	_, _ = tt.Write([]byte{0x1, 0x2})

	want := []byte{
		1,
		0x03, 0x02,
		0x07, 0x06, 0x05, 0x04,
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
		0, 1,
		0x1,
		0,
		0x1, 0x2,
	}
	if !reflect.DeepEqual(tt.Data, want) {
		t.Errorf("Data = %#v, want %#v", tt.Data, want)
	}
	wantS := []string{
		"foo",
	}
	if !reflect.DeepEqual(tt.Strings, wantS) {
		t.Errorf("Strings = %#v, want %#v", tt.Strings, wantS)
	}
}
