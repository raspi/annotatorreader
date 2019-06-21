package annotatorreader

import "reflect"

type DebugInfoMap map[Offset]DebugInformation

type DebugInformation struct {
	Name     string
	Size     uint64
	Kind     reflect.Kind
	Type     reflect.Type
	Value    *interface{}
	LastKind reflect.Kind
}
