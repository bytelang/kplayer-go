package module

import (
    "encoding/json"
    "fmt"
    "github.com/bytelang/kplayer/types"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/spf13/cobra"
    "sync"
)

type KeeperContext struct {
    id        string
    action    kpproto.EventAction
    ch        chan []byte
    validator func(msg []byte) bool
}

func NewKeeperContext(id string, action kpproto.EventAction, validator func(msg []byte) bool) KeeperContext {
    return KeeperContext{
        id:        id,
        action:    action,
        ch:        make(chan []byte),
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
    InitConfig(ctx *types.ClientContext, cfg json.RawMessage) error
    ValidateConfig() error
}

type ModuleManager map[string]AppModule

func NewModuleManager(modules ...AppModule) ModuleManager {
    moduleMap := make(ModuleManager)
    for _, module := range modules {
        moduleMap[module.GetModuleName()] = module
    }

    return moduleMap
}

func (mm *ModuleManager) GetModule(name string) *AppModule {
    m := (*mm)[name]
    return &m
}
