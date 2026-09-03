package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"prodApp/internal/booking"
	"prodApp/internal/config"
	"prodApp/internal/platform/auth"
	"prodApp/internal/platform/db"
	"prodApp/internal/training"
	"prodApp/internal/user"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	_ "prodApp/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           ProdApp API
// @version         1.0
// @description     Бэкенд для бронирования тренировок
// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите токен в формате: Bearer {token}

// @securityDefinitions.oauth2.accessCode  OAuth2AccessCode
// @tokenUrl                               http://localhost:8081/realms/prodApp/protocol/openid-connect/token
// @authorizationUrl                       http://localhost:8081/realms/prodApp/protocol/openid-connect/auth
// @scope.openid                           openid scope

func main() {
	log.Printf("start main")

	cfg, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	err = db.RunMigrations(cfg.DSN())

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DSN())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	psqlTrainingRepo := training.NewPostgresTrainingRepository(pool)
	trainingService := training.NewTrainingService(psqlTrainingRepo)
	trainingHandler := training.NewTrainingHandler(trainingService)

	psqlUserRepo := user.NewPostgresUserRepository(pool)
	psqlIdentityRepo := user.NewPostgresIdentityRepository(pool)
	psqlLinkCodeRepo := user.NewPostgresLinkCodeRepository(pool)
	userService := user.NewUserService(psqlUserRepo, psqlIdentityRepo, psqlLinkCodeRepo)
	userHandler := user.NewUserHandler(userService)

	psqlBookingRepo := booking.NewPostgresBookingRepository(pool)
	bookingService := booking.NewBookingService(psqlBookingRepo, psqlTrainingRepo, psqlUserRepo)
	bookingHandler := booking.NewBookingHandler(bookingService)

	tgAuth := auth.NewTelegramAuthenticator(cfg.TelegramBotToken)
	vkAuth := auth.NewVKAuthenticator(cfg.VKAppSecretKey)
	tokenIssuer := auth.NewTokenIssuer(cfg.JWTSigningKey)
	authMiddleware := auth.AuthMiddleware(tokenIssuer)

	authHandler := auth.NewAuthHandler(tgAuth, vkAuth, tokenIssuer, userService)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://vk.com", "https://web.telegram.org"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
	))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/telegram", authHandler.TelegramLogin)
			r.Post("/vk", authHandler.VKLogin)
			r.Post("/telegram/link", authHandler.TelegramLink)
			r.Post("/vk/link", authHandler.VKLink)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Route("/trainings", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Post("/create", trainingHandler.CreateTraining)
				})
				r.Get("/get/{id}", trainingHandler.GetTrainingById)
				r.Get("/get", trainingHandler.GetAllTrainings)
			})
			r.Route("/booking", func(r chi.Router) {
				r.Get("/getByUserId", bookingHandler.GetBookingsByUserId)
				r.Get("/getByTrainingId", bookingHandler.GetParticipants)
				r.Post("/signUp", bookingHandler.SignupForTraining)
				r.Post("/cancel", bookingHandler.CancelBooking)
			})
			r.Route("/user", func(r chi.Router) {
				r.Post("/linkCode", userHandler.GenerateLinkCode)

				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Get("/get/{id}", userHandler.GetUserById)
					r.Get("/get", userHandler.GetAllUsers)
					r.Post("/approve", userHandler.ApproveUser)
					r.Post("/unapprove", userHandler.UnapproveUser)
					r.Post("/makeAdmin", userHandler.MakeAdmin)
					r.Post("/removeAdmin", userHandler.RemoveAdmin)
					r.Post("/delete", userHandler.DeleteUser)
				})
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("start server")
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)

	if err != nil {
		panic(err)
	}
}
