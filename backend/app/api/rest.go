package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/rest/logger"

	"github.com/pavkazzz/cocos/backend/app/store"
	"github.com/pavkazzz/cocos/backend/app/store/service"
)

const hardBodyLimit = 1024 * 64 // limit size of body

// Rest is a rest access server
type Rest struct {
	Version string

	DataService *service.DataStore
	httpServer  *http.Server

	lock         sync.Mutex
	SharedSecret string

	pubRest public
}

// Shutdown rest http server
func (s *Rest) Shutdown() {
	log.Print("[WARN] shutdown rest server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.lock.Lock()
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[DEBUG] http shutdown error, %s", err)
		}
		log.Print("[DEBUG] shutdown http server completed")
	}

	s.lock.Unlock()
}

func (s *Rest) controllerGroups() public {

	pubGrp := public{
		dataService: s.DataService,
	}

	return pubGrp
}

func (s *Rest) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Throttle(1000), middleware.RealIP, Recoverer(log.Default()))
	router.Use(AppInfo("cocos", "pavkazzz", s.Version), Ping)

	ipFn := func(ip string) string { return store.HashValue(ip, s.SharedSecret)[:12] } // logger uses it for anonymization
	logInfoWithBody := logger.New(logger.Log(log.Default()), logger.WithBody, logger.IPfn(ipFn), logger.Prefix("[INFO]")).Handler

	s.pubRest = s.controllerGroups()
	// authHandler := s.Authenticator.Handlers()

	// router.Group(func(r chi.Router) {
	// 	r.Use(middleware.Timeout(5 * time.Second))
	// 	r.Use(logInfoWithBody, tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)), middleware.NoCache)
	// 	r.Mount("/auth", authHandler)
	// })

	// authMiddleware := s.Authenticator.Middleware()

	// public api routes
	router.Route("/api/v1", func(ropen chi.Router) {
		ropen.Use(middleware.Timeout(30 * time.Second))
		ropen.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
		ropen.Use(middleware.NoCache, logInfoWithBody)

		// ropen.Get("/coctails", s.pubRest.Cocktails)

		ropen.Post("/ingredients", s.pubRest.createIngredientsCtrl)
		ropen.Get("/ingredients", s.pubRest.getIngredientListCtrl)
	})

	// user api routes

	// admin api routes

	return router
}

// Run the lister and request's router, activate rest server
func (s *Rest) Run(port int) {
	log.Printf("[INFO] activate http rest server on port %d", port)

	s.lock.Lock()
	s.httpServer = s.makeHTTPServer(port, s.routes())
	s.httpServer.ErrorLog = log.ToStdLogger(log.Default(), "WARN")
	s.lock.Unlock()

	err := s.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)
}

func (s *Rest) makeHTTPServer(port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}
