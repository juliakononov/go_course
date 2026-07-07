package core

import "context"

type Searcher interface {
	Search(context.Context, string, int) ([]ComicsInfo, error)
}

type DB interface {
	GetAll(context.Context) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
