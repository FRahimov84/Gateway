package app

import (
	"context"
	"errors"
	"github.com/FRahimov84/Gateway/pkg/core/auth"
	"github.com/FRahimov84/Gateway/pkg/core/file"
	"github.com/FRahimov84/Gateway/pkg/core/product"
	"github.com/FRahimov84/Gateway/pkg/core/purshase"
	"github.com/FRahimov84/Gateway/pkg/core/utils"
	"github.com/FRahimov84/Mux/pkg/mux"
	"github.com/FRahimov84/myJwt/pkg/jwt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

type Server struct {
	router     *mux.ExactMux
	secret     jwt.Secret
	authClient *auth.Client
	productSvc *product.Product
	fileSvc    *file.File
	purSvc     *purshase.Purchase
}

func NewServer(router *mux.ExactMux, secret jwt.Secret, authClient *auth.Client, productSvc *product.Product, fileSvc *file.File, purSvc *purshase.Purchase) *Server {
	return &Server{router: router, secret: secret, authClient: authClient, productSvc: productSvc, fileSvc: fileSvc, purSvc: purSvc}
}

func (s *Server) Start() {
	s.InitRoutes()
}

func (s *Server) Stop() {
	// TODO: make server stop
}

type ErrorDTO struct {
	Errors []string `json:"errors"`
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleFrontPage() http.HandlerFunc {
	// executes in one goroutine
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "index.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		// executes in many goroutines
		// TODO: fetch data from multiple upstream services
		err := tpl.Execute(writer, struct{}{})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleLoginPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "login.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			Error string
		}{
			Error: "",
		}
		err := tpl.Execute(writer, data)
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "login.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			Error string
		}{
			Error: "",
		}
		err = request.ParseForm()
		if err != nil {
			data.Error = err.Error()
			err = tpl.Execute(writer, data)
			if err != nil {
				log.Printf("error while executing template %s %v", tpl.Name(), err)
			}
			log.Printf("error while parse login form: %v", err)
			return
		}

		// validation should always be on backend
		login := request.PostFormValue("login")
		if login == "" {
			data.Error = "login can't be empty"
			err = tpl.Execute(writer, data)
			if err != nil {
				log.Printf("error while executing template %s %v", tpl.Name(), err)
			}
			log.Print("login can't be empty")
			return
		}
		password := request.PostFormValue("password")
		if password == "" {
			data.Error = "password can't be empty"
			err = tpl.Execute(writer, data)
			if err != nil {
				log.Printf("error while executing template %s %v", tpl.Name(), err)
			}
			log.Print("password can't be empty")
			return
		}

		ctx, _ := context.WithTimeout(request.Context(), 3*time.Second)
		request = request.WithContext(ctx)

		token, err := s.authClient.Login(request.Context(), login, password)
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				data.Error = "auth service didn't response in given time"
				err = tpl.Execute(writer, data)
				if err != nil {
					log.Printf("error while executing template %s %v", tpl.Name(), err)
				}
				log.Print("auth service didn't response in given time")
				log.Print("another err") // parse it
				return
			case errors.Is(err, context.Canceled):
				data.Error = "auth service didn't response in given time"
				err = tpl.Execute(writer, data)
				if err != nil {
					log.Printf("error while executing template %s %v", tpl.Name(), err)
				}

				log.Print("auth service didn't response in given time")
				log.Print("another err") // parse it
				return
			case errors.Is(err, auth.ErrResponse):
				var typedErr *auth.ErrorResponse
				ok := errors.As(err, &typedErr)
				if ok {

					if utils.StringInSlice("err.password_mismatch", typedErr.Errors) {
						data.Error = "err.password_mismatch"
					}
					err := tpl.Execute(writer, data)
					if err != nil {
						log.Print(err)
					}
					return
				}
			}
			return
		}

		cookie := &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
		}
		http.SetCookie(writer, cookie)
		http.Redirect(writer, request, "/products", http.StatusMovedPermanently)
	}
}

func (s *Server) handleProductPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "authorized.gohtml"))
	if err != nil {
		panic(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			ProductError string
			List         []product.ResponseProduct
			User         auth.UserResponseDTO
		}{}
		ctx, _ := context.WithTimeout(request.Context(), 10*time.Second)
		request = request.WithContext(ctx)
		cookie, err := request.Cookie("token")
		if err != nil {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
			return
		}
		list, err := s.productSvc.ProductList(request.Context(), cookie.Value)
		if err != nil {
			data.ProductError = err.Error()
		} else {
			if list == nil {
				data.ProductError = "NO content here"
			} else {
				data.List = list
			}
		}
		profile, err := s.authClient.Profile(request.Context(), cookie.Value)
		if err != nil {
			log.Print(err)
		} else {
			data.User = profile
		}
		err = tpl.Execute(writer, data)
		if err != nil {
			log.Print(err)
		}
		return
	}
}

func (s *Server) handlePostEditPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "editProduct.gohtml"))
	if err != nil {
		panic(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			User auth.UserResponseDTO
			ID   int
		}{}
		fromContext, ok := mux.FromContext(request.Context(), "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(fromContext)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if id >= 0 {
			data.ID = id
		} else {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		cookie, err := request.Cookie("token")
		profile, err := s.authClient.Profile(request.Context(), cookie.Value)
		if err != nil {
			log.Print(err)
		} else {
			data.User = profile
		}
		err = tpl.Execute(writer, data)
		log.Print(err)
	}
}

func (s *Server) handlePostEdit() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fromContext, ok := mux.FromContext(request.Context(), "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(fromContext)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		ctx, _ := context.WithTimeout(request.Context(), 5*time.Second)
		request = request.WithContext(ctx)
		cookie, err := request.Cookie("token")
		if err != nil {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
			return
		}
		prod := product.ResponseProduct{}
		prod.Name = request.FormValue("name")
		prod.Price, err = strconv.Atoi(request.FormValue("price"))
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		prod.Description = request.FormValue("description")
		err = request.ParseForm()
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		file, head, err := request.FormFile("file")
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		defer file.Close()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		prod.Pic, err = s.fileSvc.Save(ctx, bytes, cookie.Value, head.Filename)

		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if id == 0 {
			err = s.productSvc.NewProduct(request.Context(), prod, cookie.Value)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		if id > 0 {
			value, err := strconv.Atoi(request.FormValue("id"))
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			err = s.productSvc.UpdateProduct(request.Context(), prod, cookie.Value, value)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(writer, request, "/products", http.StatusMovedPermanently)
	}
}

func (s *Server) handleDelete() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx, _ := context.WithTimeout(request.Context(), 5*time.Second)
		request = request.WithContext(ctx)
		cookie, err := request.Cookie("token")
		if err != nil {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
			return
		}
		id, err := strconv.Atoi(request.FormValue("id"))
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = s.productSvc.RemoveByID(request.Context(), cookie.Value, id)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleExit() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		cookie := &http.Cookie{
			Name:    "token",
			Value:   "",
			Expires: time.Time{},
		}

		http.SetCookie(writer, cookie)
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleMedia() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fromContext, ok := mux.FromContext(request.Context(), "url")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		ctx, _ := context.WithTimeout(request.Context(), 5*time.Hour)
		request = request.WithContext(ctx)
		cookie, err := request.Cookie("token")
		if err != nil {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
			return
		}
		bytes, err := s.fileSvc.Serve(request.Context(), fromContext, cookie.Value)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "image/jpeg")
		_, err = writer.Write(bytes)
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) handlePurchasePage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "purchase.gohtml"))
	if err != nil {
		panic(err)
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		data := struct {
			ProductError string
			List         []purshase.PurchaseDto
			User         auth.UserResponseDTO
		}{}
		ctx, _ := context.WithTimeout(request.Context(), 7*time.Second)
		request = request.WithContext(ctx)
		cookie, err := request.Cookie("token")
		if err != nil {
			http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
			return
		}
		profile, err := s.authClient.Profile(request.Context(), cookie.Value)
		if err != nil {
			log.Print(err)
		} else {
			data.User = profile
		}
		list, err := s.purSvc.PurchaseList(request.Context(), cookie.Value, profile.Id)
		if err != nil {
			data.ProductError = err.Error()
		} else {
			if list == nil {
				data.ProductError = "NO content here"
			} else {
				data.List = list
			}
		}

		err = tpl.Execute(writer, data)
		if err != nil {
			log.Print(err)
		}
		return
	}
}