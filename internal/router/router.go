package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

type router struct {
	muxRouter *mux.Router
}

func New() *router {
	return &router{muxRouter: mux.NewRouter()}
}

func (r *router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
	return r.muxRouter.HandleFunc(path, f)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.muxRouter.ServeHTTP(w, req)
}

func (r *router) Router() http.Handler {
	return r.muxRouter
}
