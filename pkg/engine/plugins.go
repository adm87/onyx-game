package engine

import (
	"fmt"
	"reflect"

	"github.com/adm87/onyx/pkg/engine/assert"
)

type Module interface {
	OnRegister(game Game)
	ID() uint64
}

type Modules interface {
	GetModuleByID(moduleID uint64) (Module, bool)
}

type modules struct {
	modules map[uint64]Module
}

func newModules() *modules {
	return &modules{
		modules: make(map[uint64]Module),
	}
}

func (m *modules) add(module Module) {
	moduleID := module.ID()
	if existingModule, exists := m.modules[moduleID]; exists {
		assert.Fatal(fmt.Errorf("module ID %d already registered to module of type %s", moduleID, reflect.TypeOf(existingModule).String()))
	}
	m.modules[moduleID] = module
}

func (m *modules) Register(game Game) {
	for _, module := range m.modules {
		module.OnRegister(game)
	}
}

func (m *modules) GetModuleByID(moduleID uint64) (Module, bool) {
	module, exists := m.modules[moduleID]
	return module, exists
}

func GetModule[T Module](game Game, moduleID uint64) T {
	if module, exists := game.Modules().GetModuleByID(moduleID); exists {
		return assert.Type[T](module)
	}
	panic(fmt.Errorf("module with ID %d not found", moduleID))
}
