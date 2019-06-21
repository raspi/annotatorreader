package annotatorreader

import "reflect"

type reflTree struct {
	Tree []reflect.Kind
}

func (rt *reflTree) getTree(v reflect.Value) {
	rt.Tree = append(rt.Tree, v.Kind())

	if v.CanInterface() {
		tmp := reflect.ValueOf(v.Interface())

		if tmp.Kind() == reflect.Array {
			rt.getTree(tmp.Index(0))
		}
	}

}
