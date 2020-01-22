package pgk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

func NewGHServer(port int) *GHServer {
	mutex := http.NewServeMux()
	return &GHServer{
		Mux:    mutex,
		Server: &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port), Handler: mutex},
		wg:     &sync.WaitGroup{},
	}
}

type GHServer struct {
	Mux    *http.ServeMux
	Server *http.Server
	wg     *sync.WaitGroup
}

func (server *GHServer) Start() {
	go func() {
		defer server.wg.Done()

		_ = server.Server.ListenAndServe()
	}()
	server.wg.Add(1)
}

func (server *GHServer) Stop() {
	if err := server.Server.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	server.wg.Wait()
}
