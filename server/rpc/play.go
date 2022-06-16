package rpc

import (
	"context"
	"github.com/bytelang/kplayer/module/play/provider"
	"github.com/bytelang/kplayer/types"
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
	result, err := s.pi.PlayDuration(context.TODO(), args)
	if err != nil {
		return err
	}
	resultBytes, err := types.MarshalProtoMessage(result)
	if err != nil {
		return err
	}
	types.UnmarshalProtoMessage(resultBytes, reply)

	return nil
}

// Stop  stop player on idle
func (s *Play) Stop(r *http.Request, args *server.PlayStopArgs, reply *server.PlayStopReply) error {
	_, err := s.pi.PlayStop(context.TODO(), args)
	if err != nil {
		return err
	}

	return nil
}

// Pause
func (s *Play) Pause(r *http.Request, args *server.PlayPauseArgs, reply *server.PlayPauseReply) error {
	_, err := s.pi.PlayPause(context.TODO(), args)
	if err != nil {
		return err
	}
	return nil
}

// Continue
func (s *Play) Continue(r *http.Request, args *server.PlayContinueArgs, reply *server.PlayContinueReply) error {
	_, err := s.pi.PlayContinue(context.TODO(), args)
	if err != nil {
		return err
	}
	return nil
}

// Skip
func (s *Play) Skip(r *http.Request, args *server.PlaySkipArgs, reply *server.PlaySkipReply) error {
	_, err := s.pi.PlaySkip(context.TODO(), args)
	if err != nil {
		return err
	}

	return nil
}

func (s *Play) Information(r *http.Request, args *server.PlayInformationArgs, reply *server.PlayInformationReply) error {
	info, err := s.pi.PlayInformation(context.TODO(), args)
	if err != nil {
		return err
	}

	msg, err := types.MarshalProtoMessage(info)
	if err != nil {
		return err
	}

	types.UnmarshalProtoMessage(msg, reply)

	return nil
}
