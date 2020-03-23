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

package converter

import (
	"github.com/yorkie-team/yorkie/api"
	"github.com/yorkie-team/yorkie/pkg/document/change"
	"github.com/yorkie-team/yorkie/pkg/document/checkpoint"
	"github.com/yorkie-team/yorkie/pkg/document/json"
	"github.com/yorkie-team/yorkie/pkg/document/key"
	"github.com/yorkie-team/yorkie/pkg/document/operation"
	"github.com/yorkie-team/yorkie/pkg/document/time"
)

// ToChangePack converts the given model format to Protobuf format.
func ToChangePack(pack *change.Pack) *api.ChangePack {
	return &api.ChangePack{
		DocumentKey: toDocumentKey(pack.DocumentKey),
		Checkpoint:  toCheckpoint(pack.Checkpoint),
		Changes:     toChanges(pack.Changes),
	}
}

func toDocumentKey(key *key.Key) *api.DocumentKey {
	return &api.DocumentKey{
		Collection: key.Collection,
		Document:   key.Document,
	}
}

func toCheckpoint(cp *checkpoint.Checkpoint) *api.Checkpoint {
	return &api.Checkpoint{
		ServerSeq: cp.ServerSeq,
		ClientSeq: cp.ClientSeq,
	}
}

func toChanges(changes []*change.Change) []*api.Change {
	var pbChanges []*api.Change
	for _, c := range changes {
		pbChanges = append(pbChanges, &api.Change{
			Id:         toChangeID(c.ID()),
			Message:    c.Message(),
			Operations: ToOperations(c.Operations()),
		})
	}

	return pbChanges
}

func toChangeID(id *change.ID) *api.ChangeID {
	return &api.ChangeID{
		ClientSeq: id.ClientSeq(),
		Lamport:   id.Lamport(),
		ActorId:   id.Actor().String(),
	}
}

// ToDocumentKeys converts the given model format to Protobuf format.
func ToDocumentKeys(keys ...*key.Key) []*api.DocumentKey {
	var pbKeys []*api.DocumentKey
	for _, k := range keys {
		pbKeys = append(pbKeys, toDocumentKey(k))
	}
	return pbKeys
}

// ToOperations converts the given model format to Protobuf format.
func ToOperations(operations []operation.Operation) []*api.Operation {
	var pbOperations []*api.Operation

	for _, o := range operations {
		pbOperation := &api.Operation{}
		switch op := o.(type) {
		case *operation.Set:
			pbOperation.Body = &api.Operation_Set_{
				Set: &api.Operation_Set{
					ParentCreatedAt: toTimeTicket(op.ParentCreatedAt()),
					Key:             op.Key(),
					Value:           toJSONElement(op.Value()),
					ExecutedAt:      toTimeTicket(op.ExecutedAt()),
				},
			}
		case *operation.Add:
			pbOperation.Body = &api.Operation_Add_{
				Add: &api.Operation_Add{
					ParentCreatedAt: toTimeTicket(op.ParentCreatedAt()),
					PrevCreatedAt:   toTimeTicket(op.PrevCreatedAt()),
					Value:           toJSONElement(op.Value()),
					ExecutedAt:      toTimeTicket(op.ExecutedAt()),
				},
			}
		case *operation.Remove:
			pbOperation.Body = &api.Operation_Remove_{
				Remove: &api.Operation_Remove{
					ParentCreatedAt: toTimeTicket(op.ParentCreatedAt()),
					CreatedAt:       toTimeTicket(op.CreatedAt()),
					ExecutedAt:      toTimeTicket(op.ExecutedAt()),
				},
			}
		case *operation.Edit:
			pbOperation.Body = &api.Operation_Edit_{
				Edit: &api.Operation_Edit{
					ParentCreatedAt:     toTimeTicket(op.ParentCreatedAt()),
					From:                toTextNodePos(op.From()),
					To:                  toTextNodePos(op.To()),
					CreatedAtMapByActor: toCreatedAtMapByActor(op.CreatedAtMapByActor()),
					Content:             op.Content(),
					ExecutedAt:          toTimeTicket(op.ExecutedAt()),
				},
			}
		case *operation.Select:
			pbOperation.Body = &api.Operation_Select_{
				Select: &api.Operation_Select{
					ParentCreatedAt: toTimeTicket(op.ParentCreatedAt()),
					From:            toTextNodePos(op.From()),
					To:              toTextNodePos(op.To()),
					ExecutedAt:      toTimeTicket(op.ExecutedAt()),
				},
			}
		default:
			panic("unsupported operation")
		}
		pbOperations = append(pbOperations, pbOperation)
	}

	return pbOperations
}

func toJSONElement(elem json.Element) *api.JSONElementSimple {
	switch elem := elem.(type) {
	case *json.Object:
		return &api.JSONElementSimple{
			Type:      api.ValueType_JSON_OBJECT,
			CreatedAt: toTimeTicket(elem.CreatedAt()),
		}
	case *json.Array:
		return &api.JSONElementSimple{
			Type:      api.ValueType_JSON_ARRAY,
			CreatedAt: toTimeTicket(elem.CreatedAt()),
		}
	case *json.Primitive:
		switch elem.ValueType() {
		case json.Boolean:
			return &api.JSONElementSimple{
				Type:      api.ValueType_BOOLEAN,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.Integer:
			return &api.JSONElementSimple{
				Type:      api.ValueType_INTEGER,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.Long:
			return &api.JSONElementSimple{
				Type:      api.ValueType_LONG,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.Double:
			return &api.JSONElementSimple{
				Type:      api.ValueType_DOUBLE,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.String:
			return &api.JSONElementSimple{
				Type:      api.ValueType_STRING,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.Bytes:
			return &api.JSONElementSimple{
				Type:      api.ValueType_BYTES,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		case json.Date:
			return &api.JSONElementSimple{
				Type:      api.ValueType_DATE,
				CreatedAt: toTimeTicket(elem.CreatedAt()),
				Value:     elem.Bytes(),
			}
		}
	case *json.Text:
		return &api.JSONElementSimple{
			Type:      api.ValueType_TEXT,
			CreatedAt: toTimeTicket(elem.CreatedAt()),
		}
	}
	panic("fail to encode JSONElement to protobuf")
}

func toTextNodePos(pos *json.TextNodePos) *api.TextNodePos {
	return &api.TextNodePos{
		CreatedAt:      toTimeTicket(pos.ID().CreatedAt()),
		Offset:         int32(pos.ID().Offset()),
		RelativeOffset: int32(pos.RelativeOffset()),
	}
}

func toCreatedAtMapByActor(
	createdAtMapByActor map[string]*time.Ticket,
) map[string]*api.TimeTicket {
	pbCreatedAtMapByActor := make(map[string]*api.TimeTicket)
	for actor, createdAt := range createdAtMapByActor {
		pbCreatedAtMapByActor[actor] = toTimeTicket(createdAt)
	}
	return pbCreatedAtMapByActor
}

func toTimeTicket(ticket *time.Ticket) *api.TimeTicket {
	return &api.TimeTicket{
		Lamport:   ticket.Lamport(),
		Delimiter: ticket.Delimiter(),
		ActorId:   ticket.ActorIDHex(),
	}
}
