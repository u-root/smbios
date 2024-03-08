// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package smbios

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

var (
	testbinary = "testdata/satellite_pro_l70_testdata.bin"
)

func checkError(got error, want error) bool {
	if got != nil && want != nil {
		if got.Error() == want.Error() {
			return true
		}
	}

	return errors.Is(got, want)
}

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

func Test64GetByteAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name         string
		offset       int
		expectedByte uint8
		want         error
	}{
		{
			name:         "GetByteAt",
			offset:       0,
			expectedByte: 1,
			want:         nil,
		},
		{
			name:         "GetByteAt Wrong Offset",
			offset:       213,
			expectedByte: 0,
			want:         fmt.Errorf("invalid offset %d", 213),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultByte, err := testStruct.GetByteAt(tt.offset)
			if !checkError(err, tt.want) {
				t.Errorf("GetByteAt(): '%v', want '%v'", err, tt.want)
			}
			if resultByte != tt.expectedByte {
				t.Errorf("GetByteAt() = %x, want %x", resultByte, tt.expectedByte)
			}
		})
	}
}

func Test64GetBytesAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name          string
		offset        int
		length        int
		expectedBytes []byte
		want          error
	}{
		{
			name:          "Get two bytes",
			offset:        0,
			length:        2,
			expectedBytes: []byte{1, 0},
			want:          nil,
		},
		{
			name:          "Wrong Offset",
			offset:        213,
			expectedBytes: []byte{},
			want:          fmt.Errorf("invalid offset 213"),
		},
		{
			name:          "Read out-of-bounds",
			offset:        7,
			length:        16,
			expectedBytes: []byte{},
			want:          fmt.Errorf("invalid offset 7"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resultBytes, err := testStruct.GetBytesAt(tt.offset, tt.length)

			if !checkError(err, tt.want) {
				t.Errorf("GetBytesAt(): '%v', want '%v'", err, tt.want)
			}
			if !reflect.DeepEqual(resultBytes, tt.expectedBytes) && err == nil {
				t.Errorf("GetBytesAt(): Wrong byte size, %x, want %x", resultBytes, tt.expectedBytes)
			}
		})
	}
}

func Test64GetWordAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 11, 12, 0, 0, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name          string
		offset        int
		expectedBytes uint16
		want          error
	}{
		{
			name:          "Get two bytes",
			offset:        0,
			expectedBytes: 1,
			want:          nil,
		},
		{
			name:          "Wrong Offset",
			offset:        213,
			expectedBytes: 0,
			want:          fmt.Errorf("invalid offset 213"),
		},
		{
			name:          "Read position 7",
			offset:        7,
			expectedBytes: 0xc0b,
			want:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resultBytes, err := testStruct.GetWordAt(tt.offset)
			if !checkError(err, tt.want) {
				t.Errorf("GetBytesAt(): '%v', want '%v'", err, tt.want)
			}
			if !reflect.DeepEqual(resultBytes, tt.expectedBytes) && err == nil {
				t.Errorf("GetBytesAt(): Wrong byte size, %x, want %x", resultBytes, tt.expectedBytes)
			}
		})
	}
}

func Test64GetDWordAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 11, 12, 13, 14, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name          string
		offset        int
		expectedBytes uint32
		want          error
	}{
		{
			name:          "Get two bytes",
			offset:        0,
			expectedBytes: 1,
			want:          nil,
		},
		{
			name:          "Wrong Offset",
			offset:        213,
			expectedBytes: 0,
			want:          fmt.Errorf("invalid offset 213"),
		},
		{
			name:          "Read position 7",
			offset:        7,
			expectedBytes: 0xe0d0c0b,
			want:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resultBytes, err := testStruct.GetDWordAt(tt.offset)
			if !checkError(err, tt.want) {
				t.Errorf("GetBytesAt(): '%v', want '%v'", err, tt.want)
			}
			if !reflect.DeepEqual(resultBytes, tt.expectedBytes) && err == nil {
				t.Errorf("GetBytesAt(): Wrong byte size, %x, want %x", resultBytes, tt.expectedBytes)
			}
		})
	}
}

func Test64GetQWordAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 11, 12, 13, 14, 15, 16, 17, 18},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name          string
		offset        int
		expectedBytes uint64
		want          error
	}{
		{
			name:          "Get two bytes",
			offset:        0,
			expectedBytes: 0xb0000d500000001,
			want:          nil,
		},
		{
			name:          "Wrong Offset",
			offset:        213,
			expectedBytes: 0,
			want:          fmt.Errorf("invalid offset 213"),
		},
		{
			name:          "Read position 7",
			offset:        7,
			expectedBytes: 0x1211100f0e0d0c0b,
			want:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultBytes, err := testStruct.GetQWordAt(tt.offset)
			if !checkError(err, tt.want) {
				t.Errorf("GetBytesAt(): '%v', want '%v'", err, tt.want)
			}
			if !reflect.DeepEqual(resultBytes, tt.expectedBytes) && err == nil {
				t.Errorf("GetBytesAt(): Wrong byte size, %x, want %x", resultBytes, tt.expectedBytes)
			}
		})
	}
}

func Test64GetStringAt(t *testing.T) {
	testStruct := Table{
		Header: Header{
			Type:   TableTypeBIOSInfo,
			Length: 16,
			Handle: 0,
		},
		Data:    []byte{1, 0, 0, 0, 213, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Strings: []string{"BIOS Boot Complete", "TestString #1"},
	}

	tests := []struct {
		name           string
		offset         int
		expectedString string
	}{
		{
			name:           "Valid offset",
			offset:         0,
			expectedString: "BIOS Boot Complete",
		},
		{
			name:           "Not Specified",
			offset:         2,
			expectedString: "Not Specified",
		},
		{
			name:           "Bad Index",
			offset:         4,
			expectedString: "<BAD INDEX>",
		},
	}

	for _, tt := range tests {
		resultString, _ := testStruct.GetStringAt(tt.offset)
		if resultString != tt.expectedString {
			t.Errorf("GetStringAt(): %s, want '%s'", resultString, tt.expectedString)
		}
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
}
