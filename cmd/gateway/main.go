package main

import (
	"flag"
	"github.com/FRahimov84/Gateway/cmd/gateway/app"
	"github.com/FRahimov84/Gateway/pkg/core/auth"
	"github.com/FRahimov84/Gateway/pkg/core/file"
	"github.com/FRahimov84/Gateway/pkg/core/product"
	"github.com/FRahimov84/Gateway/pkg/core/purshase"
	"github.com/FRahimov84/Mux/pkg/mux"
	"github.com/FRahimov84/di/pkg/di"
	"github.com/FRahimov84/myJwt/pkg/jwt"
	"net"
	"net/http"
	"os"
)

var (
	host = flag.String("host", "", "Server host")
	port = flag.String("port", "", "Server port")
	authUrl = flag.String("authUrl", "https://auth-servisss.herokuapp.com", "Auth Service URL")
	prodUrl = flag.String("prodUrl", "https://product-servisss.herokuapp.com", "Product Service URL")
	fileUrl = flag.String("fileUrl", "http://localhost:9888", "File Service URL")
	purUrl = flag.String("purUrl", "https://purchase-servisss.herokuapp.com", "Purchase Service URL")
//	dsn  = flag.String("dsn", "", "Postgres DSN")
)

//-host 0.0.0.0 -port 9999 -dsn postgres://user:pass@localhost:5432/auth
const (
	envHost = "HOST"
	envPort = "PORT"
	//envDSN  = "DATABASE_URL"
)

type DSN string

func main() {
	flag.Parse()
	serverHost := checkENV(envHost, *host)
	serverPort := checkENV(envPort, *port)
//	serverDsn := checkENV(envDSN, *dsn)
	addr := net.JoinHostPort(serverHost, serverPort)
	secret := jwt.Secret("secret")
	start(addr, secret, auth.Url(*authUrl), product.Url(*prodUrl), file.Url(*fileUrl), purshase.Url(*purUrl))
}

func checkENV(env string, loc string) string {
	str, ok := os.LookupEnv(env)
	if !ok {
		return loc
	}
	return str
}

func start(addr string, secret jwt.Secret, authUrl auth.Url, prodUrl product.Url, fileUrl file.Url, purUrl purshase.Url) {
	container := di.NewContainer()
	container.Provide(
		app.NewServer,
		mux.NewExactMux,
		auth.NewClient,
		file.NewFile,
		product.NewProduct,
		purshase.NewPurchase,
		func() jwt.Secret { return secret },
		func() auth.Url { return authUrl },
		func() product.Url { return prodUrl },
		func() file.Url { return fileUrl },
		func() purshase.Url { return purUrl },
	)

	container.Start()

	var appServer *app.Server
	container.Component(&appServer)
	panic(http.ListenAndServe(addr, appServer))
}
