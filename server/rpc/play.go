package rpc

import (
    "github.com/bytelang/kplayer/module/play/provider"
    "net/http"

    "github.com/bytelang/kplayer/types/server"
)

// Play rpc
type Play struct {
    pi provider.ProviderI
}

func NewPlay(pi provider.ProviderI) *Play {
    return &Play{pi: pi}
}

// Stop  stop player on idle
func (s *Play) Stop(r *http.Request, args *server.PlayStopArgs, reply *server.PlayStopReply) error {
    _, err := s.pi.PlayStop(args)
    if err != nil {
        return err
    }

    return nil
}
