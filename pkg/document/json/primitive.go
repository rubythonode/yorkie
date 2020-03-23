/*
 * Copyright 2020 The Yorkie Authors. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package json

import (
	"encoding/binary"
	"fmt"
	"math"
	time2 "time"

	"github.com/yorkie-team/yorkie/pkg/document/time"
)

type ValueType int

const (
	Null    ValueType = 0
	Boolean ValueType = 1
	Integer ValueType = 2
	Long    ValueType = 3
	Double  ValueType = 4
	String  ValueType = 5
	Bytes   ValueType = 6
	Date    ValueType = 7
)

// ValueFromBytes parses the given bytes into value.
func ValueFromBytes(valueType ValueType, value []byte) interface{} {
	switch valueType {
	case Boolean:
		if value[0] == 1 {
			return true
		}
		return false
	case Integer:
		val := int32(binary.LittleEndian.Uint32(value))
		return int(val)
	case Long:
		return int64(binary.LittleEndian.Uint64(value))
	case Double:
		return math.Float64frombits(binary.LittleEndian.Uint64(value))
	case String:
		return string(value)
	case Bytes:
		return value
	case Date:
		v := int64(binary.LittleEndian.Uint64(value))
		return time2.Unix(v, 0)
	}

	panic("unsupported type")
}

// Primitive represents JSON primitive data type including logical lock.
type Primitive struct {
	valueType ValueType
	value     interface{}
	createdAt *time.Ticket
	deletedAt *time.Ticket
}

// NewPrimitive creates a new instance of Primitive.
func NewPrimitive(value interface{}, createdAt *time.Ticket) *Primitive {
	switch val := value.(type) {
	case bool:
		return &Primitive{
			valueType: Boolean,
			value:     val,
			createdAt: createdAt,
		}
	case int:
		return &Primitive{
			valueType: Integer,
			value:     val,
			createdAt: createdAt,
		}
	case int64:
		return &Primitive{
			valueType: Long,
			value:     val,
			createdAt: createdAt,
		}
	case float64:
		return &Primitive{
			valueType: Double,
			value:     val,
			createdAt: createdAt,
		}
	case string:
		return &Primitive{
			valueType: String,
			value:     val,
			createdAt: createdAt,
		}
	case []byte:
		return &Primitive{
			valueType: Bytes,
			value:     val,
			createdAt: createdAt,
		}
	case time2.Time:
		return &Primitive{
			valueType: Date,
			value:     val,
			createdAt: createdAt,
		}
	}

	panic("unsupported type")
}

// Bytes creates an array representing the value.
func (p *Primitive) Bytes() []byte {
	switch val := p.value.(type) {
	case bool:
		if val {
			return []byte{1}
		}
		return []byte{0}
	case int:
		bytes := [4]byte{}
		binary.LittleEndian.PutUint32(bytes[:], uint32(val))
		return bytes[:]
	case int64:
		bytes := [8]byte{}
		binary.LittleEndian.PutUint64(bytes[:], uint64(val))
		return bytes[:]
	case float64:
		bytes := [8]byte{}
		binary.LittleEndian.PutUint64(bytes[:], math.Float64bits(val))
		return bytes[:]
	case string:
		return []byte(val)
	case []byte:
		return val
	case time2.Time:
		bytes := [8]byte{}
		binary.LittleEndian.PutUint64(bytes[:], uint64(val.UTC().Unix()))
		return bytes[:]
	}

	panic("unsupported type")
}

// Marshal returns the JSON encoding of the value.
func (p *Primitive) Marshal() string {
	switch p.valueType {
	case Boolean:
		return fmt.Sprintf("%t", p.value)
	case Integer:
		return fmt.Sprintf("%d", p.value)
	case Long:
		return fmt.Sprintf("%d", p.value)
	case Double:
		return fmt.Sprintf("%f", p.value)
	case String:
		return fmt.Sprintf("\"%s\"", p.value)
	case Bytes:
		// TODO: JSON.stringify({a: new Uint8Array([1,2]), b: 2})
		// {"a":{"0":1,"1":2},"b":2}
		return fmt.Sprintf("\"%s\"", p.value)
	case Date:
		return p.value.(time2.Time).Format(time2.RFC3339)
	}

	panic("unsupported type")
}

// DeepCopy copies itself deeply.
func (p *Primitive) DeepCopy() Element {
	return p
}

// CreatedAt returns the creation time.
func (p *Primitive) CreatedAt() *time.Ticket {
	return p.createdAt
}

// DeletedAt returns the deletion time of this element.
func (p *Primitive) DeletedAt() *time.Ticket {
	return p.deletedAt
}

// Delete deletes this element.
func (p *Primitive) Delete(deletedAt *time.Ticket) {
	p.deletedAt = deletedAt
}

// ValueType returns the type of the value.
func (p *Primitive) ValueType() ValueType {
	return p.valueType
}
