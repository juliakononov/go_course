package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	"github.com/dustinkirkland/golang-petname"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	petnamepb "yadro.com/course/proto"
)

type Config struct {
	Port int `yaml:"port" env:"PETNAME_GRPC_PORT" env-default:"8080"`
}

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Generate(_ context.Context, req *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	if req.Words <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "words count must be > 0")
	}

	return &petnamepb.PetnameResponse{Name: petname.Generate(int(req.Words), req.Separator)}, nil
}

func (s *server) GenerateMany(req *petnamepb.PetnameStreamRequest, stream grpc.ServerStreamingServer[petnamepb.PetnameResponse]) error {
	if req.Words <= 0 || req.Names <= 0 {
		return status.Errorf(codes.InvalidArgument, "words and names count must be > 0")
	}

	for range req.Names {
		if err := stream.Send(&petnamepb.PetnameResponse{Name: petname.Generate(int(req.Words), req.Separator)}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			slog.Error("petname: cannot read config", "error", err)
			return
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			slog.Error("petname: cannot read env", "error", err)
			return
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("petname: failed to listen", "error", err)
		return
	}

	slog.Info("petname: starting server", "addr", addr)

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		slog.Error("petname: failed to serve", "error", err)
		return
	}
}
