package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log.With("component", "update-adapter"),
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

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	reply, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("status failed", "error", err)
		return core.StatusUpdateUnknown, core.ErrInternal
	}
	switch reply.Status {
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle, nil
	case updatepb.Status_STATUS_RUNNING:
		return core.StatusUpdateRunning, nil
	default:
		return core.StatusUpdateUnknown, nil
	}
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	stats, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("stats failed", "error", err)
		return core.UpdateStats{}, core.ErrInternal
	}

	return core.UpdateStats{
		WordsTotal:    int(stats.WordsTotal),
		WordsUnique:   int(stats.WordsUnique),
		ComicsTotal:   int(stats.ComicsTotal),
		ComicsFetched: int(stats.ComicsFetched),
	}, nil

}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("update failed", "error", err)
		switch status.Code(err) {
		case codes.AlreadyExists:
			return core.ErrAlreadyExists
		default:
			return core.ErrInternal
		}
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("drop failed", "error", err)
		return core.ErrInternal
	}
	return nil
}
