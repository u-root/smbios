// Copyright 2016-2019 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dmidecode

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/u-root/smbios"
)

// Generic errors.
var (
	ErrInvalidArg = errors.New("invalid argument")
)

// We need this for testing.
type parseStructure func(t *smbios.Table, off int, complete bool, sp interface{}) (int, error)

type fieldParser interface {
	ParseField(t *smbios.Table, off int) (int, error)
}

type fieldWriter interface {
	WriteField(t *smbios.Table) (int, error)
}

var (
	fieldTagKey              = "smbios" // Tag key for annotations.
	fieldParserInterfaceType = reflect.TypeOf((*fieldParser)(nil)).Elem()
	fieldWriterInterfaceType = reflect.TypeOf((*fieldWriter)(nil)).Elem()
)

func parseStruct(t *smbios.Table, off int, complete bool, sp interface{}) (int, error) {
	var err error
	var ok bool
	var sv reflect.Value
	if sv, ok = sp.(reflect.Value); !ok {
		sv = reflect.Indirect(reflect.ValueOf(sp)) // must be a pointer to struct then, dereference it
	}
	svtn := sv.Type().Name()
	// fmt.Printf("t %s\n", svtn)
	i := 0
	for ; i < sv.NumField() && off < len(t.Data); i++ {
		f := sv.Type().Field(i)
		fv := sv.Field(i)
		ft := fv.Type()
		tags := f.Tag.Get(fieldTagKey)
		// fmt.Printf("XX %02Xh f %s t %s k %s %s\n", off, f.Name, f.Type.Name(), fv.Kind(), tags)
		// Check tags first
		ignore := false
		for _, tag := range strings.Split(tags, ",") {
			tp := strings.Split(tag, "=")
			switch tp[0] {
			case "-":
				ignore = true
			}
		}
		if ignore {
			continue
		}
		var verr error
		switch fv.Kind() {
		case reflect.Uint8:
			v, err := t.GetByteAt(off)
			fv.SetUint(uint64(v))
			verr = err
			off++
		case reflect.Uint16:
			v, err := t.GetWordAt(off)
			fv.SetUint(uint64(v))
			verr = err
			off += 2
		case reflect.Uint32:
			v, err := t.GetDWordAt(off)
			fv.SetUint(uint64(v))
			verr = err
			off += 4
		case reflect.Uint64:
			v, err := t.GetQWordAt(off)
			fv.SetUint(v)
			verr = err
			off += 8
		case reflect.String:
			v, _ := t.GetStringAt(off)
			fv.SetString(v)
			off++
		case reflect.Array:
			v, err := t.GetBytesAt(off, fv.Len())
			reflect.Copy(fv, reflect.ValueOf(v))
			verr = err
			off += fv.Len()
		default:
			if reflect.PtrTo(ft).Implements(fieldParserInterfaceType) {
				off, err = fv.Addr().Interface().(fieldParser).ParseField(t, off)
				if err != nil {
					return off, err
				}
				break
			}
			// If it's a struct, just invoke parseStruct recursively.
			if fv.Kind() == reflect.Struct {
				off, err = parseStruct(t, off, true /* complete */, fv)
				if err != nil {
					return off, err
				}
				break
			}
			return off, fmt.Errorf("%s.%s: unsupported type %s", svtn, f.Name, fv.Kind())
		}
		if verr != nil {
			return off, fmt.Errorf("failed to parse %s.%s: %w", svtn, f.Name, verr)
		}
	}

	if complete && i < sv.NumField() {
		return off, fmt.Errorf("%w: %s incomplete, got %d of %d fields", io.ErrUnexpectedEOF, svtn, i, sv.NumField())
	}

	// Fill in defaults
	for ; i < sv.NumField(); i++ {
		f := sv.Type().Field(i)
		fv := sv.Field(i)
		ft := fv.Type()
		tags := f.Tag.Get(fieldTagKey)
		// fmt.Printf("XX %02Xh f %s t %s k %s %s\n", off, f.Name, f.Type.Name(), fv.Kind(), tags)
		// Check tags first
		ignore := false
		var defValue uint64
		for _, tag := range strings.Split(tags, ",") {
			tp := strings.Split(tag, "=")
			switch tp[0] {
			case "-":
				ignore = true
			case "default":
				defValue, _ = strconv.ParseUint(tp[1], 0, 64)
			}
		}
		if ignore {
			continue
		}
		switch fv.Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fv.SetUint(defValue)
			off += int(ft.Size())

		case reflect.Array:
			return off, fmt.Errorf("%w: array does not support default values", ErrInvalidArg)

		case reflect.Struct:
			off, err := parseStruct(t, off, false /* complete */, fv)
			if err != nil {
				return off, err
			}
		}
	}

	return off, nil
}

// Table is an SMBIOS table.
type Table interface {
	Typ() smbios.TableType
}

const headerLen = 4

// ToTable converts a struct to an smbios.Table.
func ToTable(val Table, handle uint16) (*smbios.Table, error) {
	t := smbios.Table{
		Header: smbios.Header{
			Type:   val.Typ(),
			Handle: handle,
		},
	}
	m, err := toTable(&t, val)
	if err != nil {
		return nil, err
	}
	if m+headerLen > math.MaxUint8 {
		return nil, fmt.Errorf("%w: table is too long, got %d bytes, max is 256", ErrInvalidArg, m+headerLen)
	}
	t.Length = uint8(m + headerLen)
	return &t, nil
}

func toTable(t *smbios.Table, sp any) (int, error) {
	sv, ok := sp.(reflect.Value)
	if !ok {
		sv = reflect.Indirect(reflect.ValueOf(sp)) // must be a pointer to struct then, dereference it
	}
	svtn := sv.Type().Name()

	n := 0
	for i := 0; i < sv.NumField(); i++ {
		f := sv.Type().Field(i)
		fv := sv.Field(i)
		ft := fv.Type()
		tags := f.Tag.Get(fieldTagKey)
		// Check tags first
		ignore := false
		for _, tag := range strings.Split(tags, ",") {
			tp := strings.Split(tag, "=")
			switch tp[0] { //nolint
			case "-":
				ignore = true
			}
		}
		if ignore {
			continue
		}
		// fieldWriter takes precedence.
		if reflect.PtrTo(ft).Implements(fieldWriterInterfaceType) {
			m, err := fv.Addr().Interface().(fieldWriter).WriteField(t)
			n += m
			if err != nil {
				return n, err
			}
			continue
		}

		var verr error
		switch fv.Kind() {
		case reflect.Uint8:
			t.WriteByte(uint8(fv.Uint()))
			n++
		case reflect.Uint16:
			t.WriteWord(uint16(fv.Uint()))
			n += 2
		case reflect.Uint32:
			t.WriteDWord(uint32(fv.Uint()))
			n += 4
		case reflect.Uint64:
			t.WriteQWord(fv.Uint())
			n += 8
		case reflect.String:
			t.WriteString(fv.String())
			n++
		case reflect.Array:
			t.WriteBytes(fv.Slice(0, fv.Len()).Bytes())
			n += fv.Len()
		case reflect.Struct:
			var m int
			// If it's a struct, just invoke toTable recursively.
			m, verr = toTable(t, fv)
			n += m
		default:
			return n, fmt.Errorf("%s.%s: unsupported type %s", svtn, f.Name, fv.Kind())
		}
		if verr != nil {
			return n, fmt.Errorf("failed to parse %s.%s: %w", svtn, f.Name, verr)
		}
	}
	return n, nil
}
