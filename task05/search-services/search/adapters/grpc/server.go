package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	reply, err := s.service.Search(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		if errors.Is(err, core.ErrBadArguments) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	comics := make([]*searchpb.ComicsInfo, 0, len(reply))
	for _, c := range reply {
		comics = append(comics, &searchpb.ComicsInfo{
			ID:  int64(c.ID),
			URL: c.URL,
		})
	}

	return &searchpb.SearchReply{Comics: comics}, nil
}
