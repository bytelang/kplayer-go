package module

import (
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/spf13/cobra"
	"sync"
)

type ModuleOption int

const (
	ModuleOptionGenerateCache ModuleOption = iota
)

type KeeperContext struct {
	id        string
	action    kpproto.EventMessageAction
	ch        chan string
	validator func(msg string) bool
}

func NewKeeperContext(id string, action kpproto.EventMessageAction, validator func(msg string) bool) KeeperContext {
	return KeeperContext{
		id:        id,
		action:    action,
		ch:        make(chan string),
		validator: validator,
	}
}

func (kc *KeeperContext) Close() {
	close(kc.ch)
}

func (kc KeeperContext) Wait() {
	_ = <-kc.ch
}

func (kc KeeperContext) GetId() string {
	return kc.id
}

type ModuleKeeper struct {
	keeper       []KeeperContext
	triggerMutex sync.Mutex
}

func (m *ModuleKeeper) GetKeeperContext(id string) *KeeperContext {
	for _, item := range m.keeper {
		if item.id == id {
			return &item
		}
	}

	return nil
}

func (m *ModuleKeeper) RegisterKeeperChannel(ctx KeeperContext) error {
	m.triggerMutex.Lock()
	defer m.triggerMutex.Unlock()
	if m.GetKeeperContext(ctx.id) != nil {
		return fmt.Errorf("id has existed: %s", ctx.id)
	}
	m.keeper = append(m.keeper, ctx)

	return nil
}

func (m *ModuleKeeper) Trigger(message *kpproto.KPMessage) {
	m.triggerMutex.Lock()
	defer m.triggerMutex.Unlock()

	for key, item := range m.keeper {
		if item.action == message.Action {
			if item.validator(message.Body) {
				item.ch <- message.Body
				m.keeper = append(m.keeper[:key], m.keeper[key+1:]...)
			}
		}
	}
}

type BasicAppModule interface {
	RegisterKeeperChannel(ctx KeeperContext) error
	GetKeeperContext(id string) *KeeperContext
	ParseMessage(message *kpproto.KPMessage)
	TriggerMessage(message *kpproto.KPMessage)
}

type AppModule interface {
	BasicAppModule
	GetModuleName() string
	GetCommand() *cobra.Command
	InitConfig(ctx *types.ClientContext, cfg json.RawMessage) (interface{}, error)
	ValidateConfig() error
	BeginRunning(...ModuleOption)
	EndRunning(...ModuleOption)
}

type ModuleManager struct {
	Modules         map[string]AppModule
	OrderInitConfig []string
}

func NewModuleManager(modules ...AppModule) ModuleManager {
	moduleMap := ModuleManager{
		Modules: make(map[string]AppModule, 0),
	}

	for _, module := range modules {
		moduleMap.Modules[module.GetModuleName()] = module
	}

	return moduleMap
}

func (mm *ModuleManager) GetModule(name string) AppModule {
	m := mm.Modules[name]
	return m
}

func (mm *ModuleManager) SetOrderInitConfig(order ...string) {
	mm.OrderInitConfig = order
}
