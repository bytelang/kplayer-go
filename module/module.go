package module

import (
    "encoding/json"
    "fmt"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/types"
    "github.com/golang/protobuf/proto"
    "github.com/spf13/cobra"
)

type KeeperContext struct {
    id     string
    action kpproto.EventAction
    ch     chan []byte
}

func NewKeeperContext(id string, action kpproto.EventAction) KeeperContext {
    return KeeperContext{
        id:     id,
        action: action,
        ch:     make(chan []byte),
    }
}

func (kc *KeeperContext) Close() {
    close(kc.ch)
}

func (kc KeeperContext) Wait(scanPtr proto.Message) error {
    d := <-kc.ch
    return proto.Unmarshal(d, scanPtr)
}

type ModuleKeeper struct {
    keeper []KeeperContext
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
    if m.GetKeeperContext(ctx.id) != nil {
        return fmt.Errorf("id has existed: %s", ctx.id)
    }
    m.keeper = append(m.keeper, ctx)

    return nil
}

func (m *ModuleKeeper) Trigger(action kpproto.EventAction, message proto.Message) {
    for key, item := range m.keeper {
        if item.action == action {
            data, err := proto.Marshal(message)
            if err != nil {
                panic(err)
            }

            item.ch <- data
        }
        m.keeper = append(m.keeper[:key], m.keeper[key+1:]...)
    }
}

type AppModule interface {
    KeeperModule
    GetModuleName() string
    GetCommand() *cobra.Command
    InitConfig(ctx types.ClientContext, data json.RawMessage)
}

type KeeperModule interface {
    RegisterKeeperChannel(ctx KeeperContext) error
    GetKeeperContext(id string) *KeeperContext
    ParseMessage(message *kpproto.KPMessage) error
}

type ModuleManager map[string]AppModule

func NewModuleManager(modules ...AppModule) ModuleManager {
    moduleMap := make(ModuleManager)
    for _, module := range modules {
        moduleMap[module.GetModuleName()] = module
    }

    return moduleMap
}
