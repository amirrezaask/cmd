package controller

import "net/http"

type RouteBuilder interface {
	MapController(prefix string, controller http.Handler)
}

func Init(router RouteBuilder) {

}
