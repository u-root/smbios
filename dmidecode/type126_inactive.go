// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"fmt"

	"github.com/u-root/smbios"
)

// Inactive table cannot be further parsed. Documentation suggests that it can be any table
// that is temporarily marked inactive by tweaking the type field.

// InactiveTable is Defined in DSP0134 7.46.
type InactiveTable struct {
	smbios.Table
}

// NewInactiveTable parses a generic Table into InactiveTable.
func NewInactiveTable(t *smbios.Table) (*InactiveTable, error) {
	if t.Type != smbios.TableTypeInactive {
		return nil, fmt.Errorf("invalid table type %d", t.Type)
	}
	return &InactiveTable{Table: *t}, nil
}

func (it *InactiveTable) String() string {
	return it.Header.String()
}
