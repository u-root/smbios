// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Table is a generic type of table that does not parsed fields,
// it only allows access to its contents by offset.
type Table struct {
	Header `smbios:"-"`

	Data    []byte   `smbios:"-"` // Structured part of the table.
	Strings []string `smbios:"-"` // Strings section.
}

// Table parsing errors.
var (
	ErrTableNotFound = errors.New("table not found")
)

// Len returns length of the structured part of the table.
func (t *Table) Len() int {
	return len(t.Data) + headerLen
}

// GetByteAt returns a byte from the structured part at the specified offset.
func (t *Table) GetByteAt(offset int) (uint8, error) {
	if offset > len(t.Data)-1 {
		return 0, fmt.Errorf("%w at offset %d", io.ErrUnexpectedEOF, offset)
	}
	return t.Data[offset], nil
}

// GetBytesAt returns a number of bytes from the structured part at the specified offset.
func (t *Table) GetBytesAt(offset, length int) ([]byte, error) {
	if offset > len(t.Data)-length {
		return nil, fmt.Errorf("%w at offset %d with length %d", io.ErrUnexpectedEOF, offset, length)
	}
	return t.Data[offset : offset+length], nil
}

// GetWordAt returns a 16-bit word from the structured part at the specified offset.
func (t *Table) GetWordAt(offset int) (res uint16, err error) {
	if offset > len(t.Data)-2 {
		return 0, fmt.Errorf("%w at offset %d with length 2", io.ErrUnexpectedEOF, offset)
	}
	err = binary.Read(bytes.NewReader(t.Data[offset:offset+2]), binary.LittleEndian, &res)
	return res, err
}

// GetDWordAt returns a 32-bit word from the structured part at the specified offset.
func (t *Table) GetDWordAt(offset int) (res uint32, err error) {
	if offset > len(t.Data)-4 {
		return 0, fmt.Errorf("%w at offset %d with length 4", io.ErrUnexpectedEOF, offset)
	}
	err = binary.Read(bytes.NewReader(t.Data[offset:offset+4]), binary.LittleEndian, &res)
	return res, err
}

// GetQWordAt returns a 64-bit word from the structured part at the specified offset.
func (t *Table) GetQWordAt(offset int) (res uint64, err error) {
	if offset > len(t.Data)-8 {
		return 0, fmt.Errorf("%w at offset %d with length 8", io.ErrUnexpectedEOF, offset)
	}
	err = binary.Read(bytes.NewReader(t.Data[offset:offset+8]), binary.LittleEndian, &res)
	return res, err
}

// GetStringAt returns a string pointed to by the byte at the specified offset in the structured part.
// NB: offset is not the string index.
func (t *Table) GetStringAt(offset int) (string, error) {
	if offset >= len(t.Data) {
		return "", fmt.Errorf("%w at offset %d with length 1", io.ErrUnexpectedEOF, offset)
	}
	stringIndex := t.Data[offset]
	switch {
	case stringIndex == 0:
		return "Not Specified", nil
	case int(stringIndex) <= len(t.Strings):
		return t.Strings[stringIndex-1], nil
	default:
		return "<BAD INDEX>", fmt.Errorf("invalid string index %d", stringIndex)
	}
}

func (t *Table) String() string {
	lines := []string{
		t.Header.String(),
		"\tHeader and Data:",
	}
	data := append(t.Header.ToBytes(), t.Data...)
	for len(data) > 0 {
		ld := data
		if len(ld) > 16 {
			ld = ld[:16]
		}
		ls := make([]string, len(ld))
		for i, d := range ld {
			ls[i] = fmt.Sprintf("%02X", d)
		}
		lines = append(lines, "\t\t"+strings.Join(ls, " "))

		data = data[len(ld):]
	}
	if len(t.Strings) > 0 {
		lines = append(lines, "\tStrings:")
		for _, s := range t.Strings {
			lines = append(lines, "\t\t"+s)
		}
	}
	return strings.Join(lines, "\n")
}

func readFormatted(r io.Reader, l int) ([]byte, error) {
	if l < 0 {
		return nil, io.ErrUnexpectedEOF
	}
	if l == 0 {
		return nil, nil
	}
	b := make([]byte, l)
	if _, err := io.ReadFull(r, b); err != nil {
		// ReadFull returns EOF only if no bytes were read, which is
		// unexpected here.
		return nil, convertUnexpectedEOF(err)
	}
	return b, nil
}

func convertUnexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func readStrings(br *bufio.Reader) ([]string, error) {
	peek, err := br.Peek(2)
	if err != nil {
		return nil, convertUnexpectedEOF(err)
	}
	if bytes.Equal(peek, []byte{0, 0}) {
		// If peek worked, this seems impossible to fail.
		_, _ = br.Discard(2)
		return nil, nil
	}

	var s []string
	for {
		b, err := br.ReadBytes(0)
		if err != nil {
			return nil, convertUnexpectedEOF(err)
		}
		b = bytes.TrimRight(b, "\x00")
		s = append(s, string(b))

		// If the next byte is another \x00, there are no more strings.
		if peek, err := br.Peek(1); err != nil {
			return nil, convertUnexpectedEOF(err)
		} else if bytes.Equal(peek, []byte{0}) {
			_, _ = br.Discard(1)
			return s, nil
		}
	}
	// unreachable.
}

// Tables is a collection of SMBIOS tables.
type Tables []*Table

// ParseTables parses all tables from a byte stream.
func ParseTables(r io.Reader) (Tables, error) {
	br := bufio.NewReader(r)
	var tt Tables
	for {
		if _, err := br.Peek(1); err == io.EOF {
			// No more data.
			return tt, nil
		}
		t, err := parseTable(br)
		if err != nil {
			return nil, err
		}
		tt = append(tt, t)
	}
	// Unreachable.
}

// TablesByType returns tables of the specified type.
//
// TablesByType is nil-safe.
func (t Tables) TablesByType(tt TableType) Tables {
	if t == nil {
		return nil
	}
	var res Tables
	for _, u := range t {
		if u.Type == tt {
			res = append(res, u)
		}
	}
	return res
}

// TableByType returns the first table of the specified type.
//
// TablesByType is nil-safe.
func (t Tables) TableByType(tt TableType) *Table {
	if t == nil {
		return nil
	}
	for _, u := range t {
		if u.Type == tt {
			return u
		}
	}
	return nil
}

// ParseTable parses a table from byte stream.
func ParseTable(r io.Reader) (*Table, error) {
	br := bufio.NewReader(r)
	return parseTable(br)
}

func parseTable(br *bufio.Reader) (*Table, error) {
	var h Header
	if err := h.Parse(br); err != nil {
		return nil, convertUnexpectedEOF(err)
	}

	// Length of formatted section is h.Length minus length of header itself.
	l := int(h.Length) - headerLen
	f, err := readFormatted(br, l)
	if err != nil {
		return nil, fmt.Errorf("reading formatted section: %w", err)
	}

	s, err := readStrings(br)
	if err != nil {
		return nil, fmt.Errorf("reading string section: %w", err)
	}
	return &Table{
		Header:  h,
		Data:    f,
		Strings: s,
	}, nil
}
