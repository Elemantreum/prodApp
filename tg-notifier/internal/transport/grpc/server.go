package grpc

import (
	"context"
	"log"
	"prodApp/proto/gen/notifierpb"

	"github.com/mr-linch/go-tg"
)

type Server struct {
	notifierpb.UnimplementedNotifierServiceServer

	client *tg.Client
}

func NewServer(token string) *Server {
	return &Server{
		client: tg.New(token),
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

	_, err := s.client.
		SendMessage(
			tg.ChatID(req.RecipientId),
			req.Message,
		).
		Do(ctx)

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
		_, err := s.client.
			SendMessage(
				tg.ChatID(id),
				req.Message,
			).
			Do(ctx)

		if err != nil {
			return nil, err
		}
	}

	return &notifierpb.NotifyResponse{}, nil
}
