// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Entry point errors.
var (
	ErrInvalidAnchor = errors.New("invalid anchor string")
)

// EntryPoint is an SMBIOS entry point.
//
// To access detailed information, use a type assertion to *Entry32 or *Entry64.
type EntryPoint interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	fmt.Stringer

	// Table returns the physical address of the table, and its size.
	Table() (address, size int)

	// Version returns the version associated with the table structure.
	Version() (major, minor, rev int)
}

func calcChecksum(data []byte, skipIndex int) uint8 {
	var cs uint8
	for i, b := range data {
		if i == skipIndex {
			continue
		}
		cs += b
	}
	return uint8(0x100 - int(cs))
}

var (
	anchor32 = []byte("_SM_")
	anchor64 = []byte("_SM3_")
)

// ParseEntry parses SMBIOS 32 or 64-bit entrypoint structure.
func ParseEntry(r io.Reader) (EntryPoint, error) {
	br := bufio.NewReader(r)
	peek, err := br.Peek(5)
	if err != nil {
		return nil, convertUnexpectedEOF(err)
	}

	var e EntryPoint
	var data []byte
	switch {
	case bytes.Equal(peek[:4], anchor32):
		data = make([]byte, 0x1f)
		e = &Entry32{}

	case bytes.Equal(peek, anchor64):
		data = make([]byte, 0x18)
		e = &Entry64{}

	default:
		return nil, fmt.Errorf("%w: %x", ErrInvalidAnchor, peek[:4])
	}

	if _, err := io.ReadFull(br, data); err != nil {
		return nil, convertUnexpectedEOF(err)
	}
	if err := e.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return e, nil
}

// Entry32 is the SMBIOS 32-Bit entry point structure, described in DSP0134 5.2.1.
type Entry32 struct {
	Anchor            [4]uint8
	Checksum          uint8
	Length            uint8
	MajorVersion      uint8
	MinorVersion      uint8
	StructMaxSize     uint16
	Revision          uint8
	Reserved          [5]uint8
	IntAnchor         [5]uint8
	IntChecksum       uint8
	StructTableLength uint16
	StructTableAddr   uint32
	NumberOfStructs   uint16
	BCDRevision       uint8
}

var _ EntryPoint = &Entry32{}

// Table returns the physical address of the table, and its size.
func (e *Entry32) Table() (address, size int) {
	return int(e.StructTableAddr), int(e.StructTableLength)
}

// Version returns the version associated with the table structure.
func (e *Entry32) Version() (major, minor, rev int) {
	return int(e.MajorVersion), int(e.MinorVersion), 0
}

// String returns a summary of the SMBIOS version.
func (e *Entry32) String() string {
	return fmt.Sprintf("SMBIOS %d.%d", e.MajorVersion, e.MinorVersion)
}

// UnmarshalBinary unmarshals the SMBIOS 32-Bit entry point structure from binary data.
func (e *Entry32) UnmarshalBinary(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, e); err != nil {
		return err
	}
	if !bytes.Equal(e.Anchor[:], anchor32) {
		return fmt.Errorf("%w: %q", ErrInvalidAnchor, string(e.Anchor[:]))
	}

	if int(e.Length) != 0x1f {
		return fmt.Errorf("length mismatch: %d vs %d", e.Length, len(data))
	}
	cs := calcChecksum(data[:e.Length], 4)
	if e.Checksum != cs {
		return fmt.Errorf("checksum mismatch: 0x%02x vs 0x%02x", e.Checksum, cs)
	}
	if !bytes.Equal(e.IntAnchor[:], []byte("_DMI_")) {
		return fmt.Errorf("invalid intermediate anchor string %q", string(e.Anchor[:]))
	}
	intCs := calcChecksum(data[0x10:0x1f], 5)
	if e.IntChecksum != intCs {
		return fmt.Errorf("intermediate checksum mismatch: 0x%02x vs 0x%02x", e.IntChecksum, intCs)
	}
	return nil
}

// MarshalBinary marshals the SMBIOS 32-Bit entry point structure to binary data.
func (e *Entry32) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, e); err != nil {
		return nil, err
	}
	// Adjust checksums.
	data := buf.Bytes()
	data[0x15] = calcChecksum(data[0x10:0x1f], 5)
	data[4] = calcChecksum(data, 4)
	return data, nil
}

// Entry64 is the SMBIOS 64-Bit entry point structure, described in DSP0134 5.2.2.
type Entry64 struct {
	Anchor          [5]uint8
	Checksum        uint8
	Length          uint8
	MajorVersion    uint8
	MinorVersion    uint8
	DocRev          uint8
	Revision        uint8
	Reserved        uint8
	StructMaxSize   uint32
	StructTableAddr uint64
}

var _ EntryPoint = &Entry64{}

// Table returns the physical address of the table, and its size.
func (e *Entry64) Table() (address, size int) {
	return int(e.StructTableAddr), int(e.StructMaxSize)
}

// Version returns the version associated with the table structure.
func (e *Entry64) Version() (major, minor, rev int) {
	return int(e.MajorVersion), int(e.MinorVersion), int(e.DocRev)
}

// String returns a summary of the SMBIOS version.
func (e *Entry64) String() string {
	return fmt.Sprintf("SMBIOS %d.%d.%d", e.MajorVersion, e.MinorVersion, e.DocRev)
}

// UnmarshalBinary unmarshals the SMBIOS 64-Bit entry point structure from binary data.
func (e *Entry64) UnmarshalBinary(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, e); err != nil {
		return err
	}
	if !bytes.Equal(e.Anchor[:], anchor64) {
		return fmt.Errorf("%w: %q", ErrInvalidAnchor, string(e.Anchor[:]))
	}

	if int(e.Length) != 0x18 {
		return fmt.Errorf("length mismatch: %d vs %d", e.Length, len(data))
	}
	cs := calcChecksum(data[:e.Length], 5)
	if e.Checksum != cs {
		return fmt.Errorf("checksum mismatch: 0x%02x vs 0x%02x", e.Checksum, cs)
	}
	return nil
}

// MarshalBinary marshals the SMBIOS 64-Bit entry point structure to binary data.
func (e *Entry64) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, e); err != nil {
		return nil, err
	}
	// Adjust checksum.
	data := buf.Bytes()
	data[5] = calcChecksum(data, 5)
	return data, nil
}
