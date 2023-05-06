// TODO: test swap corresponding types (e.g. u1 <-> u2 and u2 <-> u1)
// TODO: test exported alias refers to something in another package -- does correspondence work then?
// TODO: CODE COVERAGE
// TODO: note that we may miss correspondences because we bail early when we compare a signature (e.g. when lengths differ; we could do up to the shorter)
// TODO: if you add an unexported method to an exposed interface, you have to check that
//		every exposed type that previously implemented the interface still does. Otherwise
//		an external assignment of the exposed type to the interface type could fail.
// TODO: check constant values: large values aren't representable by some types.
// TODO: Document all the incompatibilities we don't check for.

package apidiff

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"strings"
)

// Changes reports on the differences between the APIs of the old and new packages.
// It classifies each difference as either compatible or incompatible (breaking.) For
// a detailed discussion of what constitutes an incompatible change, see the package
// documentation.
func Changes(old, new *types.Package) Report {
	d := newDiffer(old, new)
	d.checkPackage()
	r := Report{}
	for _, m := range d.incompatibles.collect() {
		r.Changes = append(r.Changes, Change{
			Message:       m.Message,
			Node:          m.Node,
			ChangedObject: m.ChangedObject,
			ChangedType:   m.ChangedType,
			Compatible:    false})
	}
	for _, m := range d.compatibles.collect() {
		r.Changes = append(r.Changes, Change{
			Message:       m.Message,
			Node:          m.Node,
			ChangedObject: m.ChangedObject,
			ChangedType:   m.ChangedType,
			Compatible:    true})
	}
	return r
}

type differ struct {
	old, new *types.Package
	// Correspondences between named types.
	// Even though it is the named types (*types.Named) that correspond, we use
	// *types.TypeName as a map key because they are canonical.
	// The values can be either named types or basic types.
	correspondMap map[*types.TypeName]types.Type

	// Messages.
	incompatibles messageSet
	compatibles   messageSet
}

func newDiffer(old, new *types.Package) *differ {
	return &differ{
		old:           old,
		new:           new,
		correspondMap: map[*types.TypeName]types.Type{},
		incompatibles: messageSet{},
		compatibles:   messageSet{},
	}
}

func (d *differ) incompatible(obj types.Object, part, changeType, format string, args ...interface{}) {
	addMessage(d.incompatibles, obj, part, changeType, format, args)
}

func (d *differ) compatible(obj types.Object, part, changeType, format string, args ...interface{}) {
	addMessage(d.compatibles, obj, part, changeType, format, args)
}

func addMessage(ms messageSet, obj types.Object, part, changeType, format string, args []interface{}) {
	ms.add(obj, part, changeMessage{
		msg:   fmt.Sprintf(format, args...),
		cType: changeType,
	})
}

func (d *differ) checkPackage() {
	// Old changes.
	for _, name := range d.old.Scope().Names() {
		oldobj := d.old.Scope().Lookup(name)
		if !oldobj.Exported() {
			continue
		}
		newobj := d.new.Scope().Lookup(name)
		if newobj == nil {
			d.incompatible(oldobj, "", typeKindString(oldobj.Type()), "removed")
			continue
		}
		d.checkObjects(oldobj, newobj)
	}
	// New additions.
	for _, name := range d.new.Scope().Names() {
		newobj := d.new.Scope().Lookup(name)
		if newobj.Exported() && d.old.Scope().Lookup(name) == nil {
			d.compatible(newobj, "", typeKindString(newobj.Type()), "added")
		}
	}

	// Whole-package satisfaction.
	// For every old exposed interface oIface and its corresponding new interface nIface...
	for otn1, nt1 := range d.correspondMap {
		oIface, ok := otn1.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}
		nIface, ok := nt1.Underlying().(*types.Interface)
		if !ok {
			// If nt1 isn't an interface but otn1 is, then that's an incompatibility that
			// we've already noticed, so there's no need to do anything here.
			continue
		}
		// For every old type that implements oIface, its corresponding new type must implement
		// nIface.
		for otn2, nt2 := range d.correspondMap {
			if otn1 == otn2 {
				continue
			}
			if types.Implements(otn2.Type(), oIface) && !types.Implements(nt2, nIface) {
				d.incompatible(otn2, "", typeKindString(nt2), "no longer implements %s", objectString(otn1))
			}
		}
	}
}

func (d *differ) checkObjects(old, new types.Object) {
	switch old := old.(type) {
	case *types.Const:
		if new, ok := new.(*types.Const); ok {
			d.constChanges(old, new)
			return
		}
	case *types.Var:
		if new, ok := new.(*types.Var); ok {
			d.checkCorrespondence(old, "", old.Type(), new.Type())
			return
		}
	case *types.Func:
		switch new := new.(type) {
		case *types.Func:
			d.checkCorrespondence(old, "", old.Type(), new.Type())
			return
		case *types.Var:
			d.compatible(old, "", typeKindString(new.Type()), "changed from func to var")
			d.checkCorrespondence(old, "", old.Type(), new.Type())
			return

		}
	case *types.TypeName:
		if new, ok := new.(*types.TypeName); ok {
			d.checkCorrespondence(old, "", old.Type(), new.Type())
			return
		}
	default:
		panic("unexpected obj type")
	}
	// Here if kind of type changed.
	d.incompatible(old, "", "typeChanged", "changed from %s to %s",
		objectKindString(old), objectKindString(new))
}

// Compare two constants.
func (d *differ) constChanges(old, new *types.Const) {
	ot := old.Type()
	nt := new.Type()
	// Check for change of type.
	if _, corr := d.correspond(ot, nt); !corr {
		d.typeChanged(old, "", "", ot, nt)
		return
	}
	// Check for change of value.
	// We know the types are the same, so constant.Compare shouldn't panic.
	if !constant.Compare(old.Val(), token.EQL, new.Val()) {
		d.incompatible(old, "", typeKindString(old.Type()), "value changed from %s to %s", old.Val(), new.Val())
	}
}

func objectKindString(obj types.Object) string {
	switch obj.(type) {
	case *types.Const:
		return "const"
	case *types.Var:
		return "var"
	case *types.Func:
		return "func"
	case *types.TypeName:
		return "type"
	default:
		return "???"
	}
}

func typeKindString(t types.Type) string {
	switch tt := t.(type) {
	case *types.Basic:
		return "basic"
	case *types.Array:
		return "array"
	case *types.Slice:
		return "slice"
	case *types.Map:
		return "map"
	case *types.Chan:
		return "chan"
	case *types.Pointer:
		return "pointer|" + typeKindString(tt.Elem())
	case *types.Signature:
		return "function"
	case *types.Tuple:
		return "tuple"
	case *types.Struct:
		return "struct"
	case *types.Interface:
		return "interface"
	case *types.Named:
		return "named|" + typeKindString(tt.Underlying())
	case *types.TypeParam:
		return "typeParam"
	default:
		return "???"
	}
}

func (d *differ) checkCorrespondence(obj types.Object, part string, old, new types.Type) {
	if msg, ok := d.correspond(old, new); !ok {
		d.typeChanged(obj, part, msg, old, new)
	}
}

func (d *differ) typeChanged(obj types.Object, part, msg string, old, new types.Type) {
	old = removeNamesFromSignature(old)
	new = removeNamesFromSignature(new)
	olds := types.TypeString(old, types.RelativeTo(d.old))
	news := types.TypeString(new, types.RelativeTo(d.new))
	m := fmt.Sprintf("%s changed from %s to %s", msg, olds, news)
	d.incompatible(obj, part, typeKindString(old), strings.TrimPrefix(m, " "))
}

// go/types always includes the argument and result names when formatting a signature.
// Since these can change without affecting compatibility, we don't want users to
// be distracted by them, so we remove them.
func removeNamesFromSignature(t types.Type) types.Type {
	sig, ok := t.(*types.Signature)
	if !ok {
		return t
	}

	dename := func(p *types.Tuple) *types.Tuple {
		var vars []*types.Var
		for i := 0; i < p.Len(); i++ {
			v := p.At(i)
			vars = append(vars, types.NewVar(v.Pos(), v.Pkg(), "", v.Type()))
		}
		return types.NewTuple(vars...)
	}

	return types.NewSignature(sig.Recv(), dename(sig.Params()), dename(sig.Results()), sig.Variadic())
}
