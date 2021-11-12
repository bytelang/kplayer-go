package rpc

import (
    "github.com/bytelang/kplayer/module/play/provider"
    "github.com/golang/protobuf/proto"
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

// Duration
func (s *Play) Duration(r *http.Request, args *server.PlayDurationArgs, reply *server.PlayDurationReply) error {
    result, err := s.pi.PlayDuration(args)
    if err != nil {
        return err
    }
    resultBytes, err := proto.Marshal(result)
    if err != nil {
        return err
    }
    if err := proto.Unmarshal(resultBytes, reply); err != nil {
        return err
    }

    return nil
}

// Stop  stop player on idle
func (s *Play) Stop(r *http.Request, args *server.PlayStopArgs, reply *server.PlayStopReply) error {
    _, err := s.pi.PlayStop(args)
    if err != nil {
        return err
    }

    return nil
}

// Pause
func (s *Play) Pause(r *http.Request, args *server.PlayPauseArgs, reply *server.PlayPauseReply) error {
    _, err := s.pi.PlayPause(args)
    if err != nil {
        return err
    }
    return nil
}

// Continue
func (s *Play) Continue(r *http.Request, args *server.PlayContinueArgs, reply *server.PlayContinueReply) error {
    _, err := s.pi.PlayContinue(args)
    if err != nil {
        return err
    }
    return nil
}

// Skip
func (s *Play) Skip(r *http.Request, args *server.PlaySkipArgs, reply *server.PlaySkipReply) error {
    _, err := s.pi.PlaySkip(args)
    if err != nil {
        return err
    }

    return nil
}

func (s *Play) Information(r *http.Request, args *server.PlayInformationArgs, reply *server.PlayInformationReply) error {
    info, err := s.pi.PlayInformation(args)
    if err != nil {
        return err
    }

    msg, err := proto.Marshal(info)
    if err != nil {
        return err
    }

    if err := proto.Unmarshal(msg, reply); err != nil {
        return err
    }

    return nil
}
