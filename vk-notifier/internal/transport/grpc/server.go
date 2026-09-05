package grpc

import (
	"context"
	"log"
	"prodApp/proto/gen/notifierpb"

	vk "github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/api/params"
)

type Server struct {
	notifierpb.UnimplementedNotifierServiceServer

	client *vk.VK
}

func NewServer(token string) *Server {
	return &Server{
		client: vk.NewVK(token),
	}
}

func (s *Server) Notify(
	ctx context.Context,
	req *notifierpb.NotifyRequest,
) (*notifierpb.NotifyResponse, error) {

	log.Printf(
		"notify recipient=%d message=%q",
		req.RecipientId,
		req.Message,
	)

	p := params.NewMessagesSendBuilder()

	p.PeerID(int(req.RecipientId))
	p.Message(req.Message)

	_, err := s.client.MessagesSend(p.Params)

	if err != nil {
		return nil, err
	}

	return &notifierpb.NotifyResponse{}, nil
}

func (s *Server) NotifyMany(
	ctx context.Context,
	req *notifierpb.NotifyManyRequest,
) (*notifierpb.NotifyResponse, error) {

	for _, id := range req.RecipientIds {
		p := params.NewMessagesSendBuilder()

		p.PeerID(int(id))
		p.Message(req.Message)

		_, err := s.client.MessagesSend(p.Params)

		if err != nil {
			return nil, err
		}
	}

	return &notifierpb.NotifyResponse{}, nil
}
