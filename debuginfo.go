package annotatorreader

import "reflect"

type DebugInfoMap map[Offset]DebugInformation

type DebugInformation struct {
	Name     string         `json:"name"`
	Size     uint64         `json:"size"`
	Type     reflect.Type   `json:"type"`
	Value    *interface{}   `json:"value,omitempty"`
	KindTree []reflect.Kind `json:"kindtree"`
}
