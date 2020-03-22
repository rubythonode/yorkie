package converter

import (
	"github.com/yorkie-team/yorkie/pkg/document/json"
	"github.com/yorkie-team/yorkie/pkg/document/time"
)

func BytesToObject(snapshot []byte) *json.Object {
	obj := json.NewObject(json.NewRHT(), time.InitialTicket)
	// TODO impl
	return obj
}