package wir

import (
	"github.com/wa-lang/wa/internal/backends/compiler_wat/wir/wat"
	"github.com/wa-lang/wa/internal/ssa"
)

/**************************************
Module:
**************************************/
type Module struct {
	Funcs []*Function

	globals     []Value
	Globals_map map[ssa.Value]Value

	BaseWat string
}

func NewModule() *Module {
	var m Module
	m.Globals_map = make(map[ssa.Value]Value)
	return &m
}

func (m *Module) AddGlobal(name string, typ ValueType, is_pointer bool, ssa_value ssa.Value) Value {
	var kind ValueKind
	if is_pointer {
		kind = ValueKindGlobal_Pointer
	} else {
		kind = ValueKindGlobal_Value
	}
	v := NewVar(name, kind, typ)
	m.globals = append(m.globals, v)
	m.Globals_map[ssa_value] = v
	return v
}

func (m *Module) ToWatModule() *wat.Module {
	var wat_module wat.Module
	wat_module.BaseWat = m.BaseWat

	for _, f := range m.Funcs {
		wat_module.Funcs = append(wat_module.Funcs, f.ToWatFunc())
	}

	for _, g := range m.globals {
		raw := g.raw()
		for _, r := range raw {
			var wat_global wat.Global
			wat_global.V = r
			wat_global.IsMut = true
			wat_module.Globals = append(wat_module.Globals, wat_global)
		}
	}

	return &wat_module
}