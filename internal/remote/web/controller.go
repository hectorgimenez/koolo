package web

import (
	"context"
	"net/http"
	"strconv"
)

type Controller struct {
	srv *http.Server
}

func New(port int) *Controller {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/action", action)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	c := &Controller{srv: srv}

	return c
}

func (c *Controller) Run() error {
	return c.srv.ListenAndServe()
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.srv.Shutdown(ctx)
}
