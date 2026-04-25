package services

import "context"

type GreetService struct {
	ctx context.Context
}

func (g *GreetService) Startup(ctx context.Context) {
	g.ctx = ctx
}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}
