// TODO: show that two-non-empty dotjoin can happen, by using an anon struct as a field type
// TODO: don't report removed/changed methods for both value and pointer method sets?

package apidiff

import (
	"fmt"
	"go/types"
	"sort"
	"strings"
)

type changeMessage struct {
	cType string
	msg   string
}

// There can be at most one message for each object or part thereof.
// Parts include interface methods and struct fields.
//
// The part thing is necessary. Method (Func) objects have sufficient info, but field
// Vars do not: they just have a field name and a type, without the enclosing struct.
type messageSet map[types.Object]map[string]changeMessage

// Add a message for obj and part, overwriting a previous message
// (shouldn't happen).
// obj is required but part can be empty.
func (m messageSet) add(obj types.Object, part string, msg changeMessage) {
	s := m[obj]
	if s == nil {
		s = map[string]changeMessage{}
		m[obj] = s
	}
	if f, ok := s[part]; ok && f != msg {
		fmt.Printf("! second, different message for obj %s, part %q\n", obj, part)
		fmt.Printf("  first:  %s\n", f)
		fmt.Printf("  second: %s\n", msg.msg)
	}
	s[part] = msg
}

func (m messageSet) collect() []Change {
	var r []Change
	for obj, parts := range m {
		// Format each object name relative to its own package.
		objstring := objectString(obj)
		for part, msg := range parts {
			var p, m string
			if strings.HasPrefix(part, ",") {
				p = objstring
				m = strings.TrimPrefix(part, ", ") + " " + msg.msg
			} else {
				p = dotjoin(objstring, part)
				m = msg.msg
			}
			r = append(r, Change{
				Node:          p,
				Message:       m,
				ChangedObject: objectKindString(obj),
				ChangedType:   msg.cType,
			})
		}
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].Node < r[j].Node
	})
	return r
}

func objectString(obj types.Object) string {
	if f, ok := obj.(*types.Func); ok {
		sig := f.Type().(*types.Signature)
		if recv := sig.Recv(); recv != nil {
			tn := types.TypeString(recv.Type(), types.RelativeTo(obj.Pkg()))
			if tn[0] == '*' {
				tn = "(" + tn + ")"
			}
			return fmt.Sprintf("%s.%s", tn, obj.Name())
		}
	}
	return obj.Name()
}

func dotjoin(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	if s2 == "" {
		return s1
	}
	return s1 + "." + s2
}
