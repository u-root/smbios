// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const headerLen = 4

// Header is the header common to all table types.
type Header struct {
	Type   TableType
	Length uint8
	Handle uint16
}

// Parse parses the header from the binary data.
func (h *Header) Parse(r io.Reader) error {
	return binary.Read(io.LimitReader(r, headerLen), binary.LittleEndian, h)
}

// ToBytes returns the header as bytes.
func (h *Header) ToBytes() []byte {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.LittleEndian, h)
	return b.Bytes()
}

// String returns string representation os the header.
func (h *Header) String() string {
	return fmt.Sprintf(
		"Handle 0x%04X, DMI type %d, %d bytes\n%s",
		h.Handle, h.Type, h.Length, h.Type)
}
