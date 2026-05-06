package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

const maxPhraseLen = 4096

type Config struct {
	Port int `yaml:"port" env:"WORDS_GRPC_PORT" env-default:"8080"`
}

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.Phrase) > maxPhraseLen {
		return nil, status.Errorf(codes.ResourceExhausted, "phrase exceeds 4096 bytes")
	}
	return &wordspb.WordsReply{
		Words: words.Norm(in.Phrase),
	}, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			slog.Error("words: cannot read config", "error", err)
			return
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			slog.Error("words: cannot read env", "error", err)
			return
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("words: failed to listen", "error", err)
		return
	}

	slog.Info("words: starting server", "addr", addr)

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		slog.Error("words: failed to serve", "error", err)
	}
}
