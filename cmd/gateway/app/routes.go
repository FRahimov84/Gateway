package app

import (
	"github.com/FRahimov84/Gateway/pkg/mux/middleware/authenticated"
	"github.com/FRahimov84/Gateway/pkg/mux/middleware/jwt"
	"github.com/FRahimov84/Gateway/pkg/mux/middleware/logger"
	"reflect"
)

func (s *Server) InitRoutes() {
	jwtMW := jwt.JWT(jwt.SourceCookie, reflect.TypeOf((*Payload)(nil)).Elem(), s.secret)
	authMW := authenticated.Authenticated(jwt.IsContextNonEmpty, true, "/")
	s.router.GET("/", s.handleFrontPage(),jwtMW, logger.Logger("HTTP"))

	// GET -> html
	s.router.GET("/login", s.handleLoginPage(),jwtMW, logger.Logger("HTTP"))
	// POST -> form handling + return HTML
	s.router.POST("/login", s.handleLogin(),jwtMW, logger.Logger("HTTP"))
	s.router.GET("/media/{url}",s.handleMedia(),authMW, jwtMW, logger.Logger("HTTP"))

	// список постов
	s.router.GET("/products", s.handleProductPage(), authMW, jwtMW, logger.Logger("HTTP"))
	s.router.GET("/purchases", s.handlePurchasePage(), authMW, jwtMW, logger.Logger("HTTP"))
	//s.router.POST("/products", s.handleProductPage(), authMW, jwtMW, logger.Logger("HTTP"))
	// форма создания/редактирования
	s.router.GET("/products/{id}/edit", s.handlePostEditPage(), authMW, jwtMW, logger.Logger("HTTP"))
	// сохранение
	s.router.POST("/products/{id}/edit", s.handlePostEdit(), authMW, jwtMW, logger.Logger("HTTP"))

	s.router.POST("/products/remove", s.handleDelete(), authMW, jwtMW, logger.Logger("HTTP"))

	s.router.GET("/exit", s.handleExit(), authMW, jwtMW, logger.Logger("HTTP"))

}
