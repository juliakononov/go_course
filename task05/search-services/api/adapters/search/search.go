package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log.With("component", "search-adapter"),
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("ping failed", "error", err)
		return err
	}
	return nil
}

func (c Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	reply, err := c.client.Search(ctx, &searchpb.SearchRequest{Phrase: phrase, Limit: int64(limit)})
	if err != nil {
		c.log.Error("search failed", "phrase_len", len(phrase), "limit", limit, "error", err)
		switch status.Code(err) {
		case codes.InvalidArgument:
			return nil, core.ErrBadArguments
		case codes.Unavailable:
			return nil, core.ErrServiceUnavailable
		default:
			return nil, core.ErrInternal
		}
	}
	res := make([]core.Comics, 0, len(reply.GetComics()))
	for _, c := range reply.GetComics() {
		res = append(res, core.Comics{ID: int(c.ID), URL: c.URL})
	}
	return res, nil
}
