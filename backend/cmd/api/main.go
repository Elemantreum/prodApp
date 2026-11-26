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
	http2 "prodApp/internal/platform/http"
	"prodApp/internal/platform/kafka"
	"prodApp/internal/platform/outbox"
	"prodApp/internal/platform/schedule"
	"prodApp/internal/section"
	"prodApp/internal/sectioninvite"
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

	outboxRepo := outbox.NewRepository(pool)

	psqlUserRepo := user.NewPostgresUserRepository(pool)
	psqlIdentityRepo := user.NewPostgresIdentityRepository(pool)
	psqlLinkCodeRepo := user.NewPostgresLinkCodeRepository(pool)

	userService := user.NewUserService(psqlUserRepo, psqlIdentityRepo, psqlLinkCodeRepo, outboxRepo)
	userHandler := user.NewUserHandler(userService)

	psqlTrainingRepo := training.NewPostgresTrainingRepository(pool)
	psqlRecurringRepo := training.NewPostgresRecurringTrainingRepository(pool)

	trainingService := training.NewTrainingService(psqlTrainingRepo, psqlRecurringRepo, outboxRepo, psqlIdentityRepo)
	trainingHandler := training.NewTrainingHandler(trainingService)

	psqlBookingRepo := booking.NewPostgresBookingRepository(pool)
	bookingService := booking.NewBookingService(psqlBookingRepo, psqlTrainingRepo, psqlUserRepo, psqlIdentityRepo)
	bookingHandler := booking.NewBookingHandler(bookingService)

	psqlSectionRepo := section.NewPostgresSectionRepository(pool)
	sectionService := section.NewSectionService(psqlSectionRepo)
	sectionAccess := section.NewSectionAccess(psqlSectionRepo)
	sectionHandler := section.NewSectionHandler(sectionService, sectionAccess)
	inviteRepo := sectioninvite.NewPostgresInvitationRepository(pool)
	inviteService := sectioninvite.NewInvitationService(inviteRepo)
	inviteHandler := sectioninvite.NewInvitationHandler(inviteService)

	tgAuth := auth.NewTelegramAuthenticator(cfg.TelegramBotToken)
	vkAuth := auth.NewVKAuthenticator(cfg.VKAppSecretKey)
	tokenIssuer := auth.NewTokenIssuer(cfg.JWTSigningKey)
	refreshTokenRepo := auth.NewRefreshTokenRepository(pool)
	authMiddleware := auth.AuthMiddleware(tokenIssuer)

	authService := auth.NewAuthService(tokenIssuer, refreshTokenRepo, userService)
	authHandler := auth.NewAuthHandler(tgAuth, vkAuth, authService, cfg.TelegramBotName)

	scheduleRepo := schedule.NewPostgresChangeRepository(cfg.DSN())
	scheduleService := schedule.NewScheduleService(scheduleRepo)
	scheduleHandler := schedule.NewScheduleHandler(scheduleService, tokenIssuer)
	updatesCtx, stopUpdates := context.WithCancel(ctx)
	defer stopUpdates()
	go scheduleService.Run(updatesCtx)
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://vk.com", "https://web.telegram.org"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "X-Section-ID"},
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		err = pool.Ping(ctx)

		if err != nil {
			http2.RespondError(w, err)
		}

		w.WriteHeader(http.StatusOK)
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
	))

	router.Route("/api/v1", func(r chi.Router) {
		r.Handle("/schedule/live", scheduleHandler)
		r.Route("/auth", func(r chi.Router) {
			r.Get("/telegram/website", authHandler.WebsiteBot)
			r.Post("/telegram/website", authHandler.WebsiteLogin)
			r.Post("/telegram", authHandler.TelegramLogin)
			r.Post("/vk", authHandler.VKLogin)
			r.Post("/telegram/link", authHandler.TelegramLink)
			r.Post("/vk/link", authHandler.VKLink)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Use(sectionAccess.Permissions(false))
			r.Get("/sections", sectionHandler.List)
			r.Post("/sections/save", sectionHandler.Save)
			r.Post("/sections/member", sectionHandler.ChangeMembership)
			r.Post("/sections/invites/preview", inviteHandler.Preview)
			r.Post("/sections/invites/redeem", inviteHandler.Redeem)
			r.With(auth.RequireAdmin).Delete("/sections", sectionHandler.Delete)
			r.Group(func(r chi.Router) {
				r.Use(sectionAccess.Permissions(true))
				r.Use(auth.RequireAdmin)
				r.Post("/sections/venues", sectionHandler.AddVenue)
				r.Delete("/sections/venues", sectionHandler.DeleteVenue)
				r.Get("/sections/invites", inviteHandler.List)
				r.Post("/sections/invites", inviteHandler.Create)
				r.Delete("/sections/invites/{id}", inviteHandler.Revoke)
			})
			r.Route("/trainings", func(r chi.Router) {
				r.Use(sectionAccess.Permissions(true))
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Post("/create", trainingHandler.CreateTraining)
					r.Post("/update", trainingHandler.UpdateTraining)
					r.Post("/delete", trainingHandler.DeleteTraining)
					r.Post("/recurring/create", trainingHandler.CreateRecurringTraining)
					r.Get("/recurring/get", trainingHandler.GetRecurringTrainings)
					r.Delete("/recurring/{id}", trainingHandler.DeleteRule)
					r.Post("/recurring/update", trainingHandler.UpdateRecurringTraining)
				})
				r.Get("/get/{id}", trainingHandler.GetTrainingById)
				r.Get("/get", trainingHandler.GetAllTrainings)
			})
			r.Route("/booking", func(r chi.Router) {
				r.Use(sectionAccess.Permissions(true))
				r.Get("/getByUserId", bookingHandler.GetBookingsByUserId)
				r.Get("/getByTrainingId", bookingHandler.GetParticipants)
				r.Get("/waitlist", bookingHandler.OwnWaitlist)
				r.Put("/waitlist", bookingHandler.SetWaiting)
				r.Delete("/waitlist", bookingHandler.SetWaiting)
				r.Get("/antiLeaderboard", bookingHandler.GetAntiLeaderboard)
				r.Post("/signUp", bookingHandler.SignupForTraining)
				r.Post("/cancel", bookingHandler.CancelBooking)
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Post("/admin/manage", bookingHandler.Manage)
					r.Get("/admin/waitlist", bookingHandler.WaitingMembers)
					r.Get("/admin/history", bookingHandler.SectionBookings)
					r.Get("/admin/members", sectionHandler.Roster)
					r.Get("/admin/candidates", sectionHandler.Candidates)
					r.Get("/admin/getByTrainingId", bookingHandler.GetTrainingBookings)
					r.Post("/setAttendance", bookingHandler.SetAttendance)
				})
			})
			r.Route("/user", func(r chi.Router) {
				r.Post("/linkCode", userHandler.GenerateLinkCode)
				r.Get("/me", userHandler.GetOwnProfile)
				r.Get("/identities", userHandler.GetOwnIdentities)
				r.Post("/updateProfile", userHandler.UpdateProfile)

				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
					r.Get("/history", bookingHandler.UserHistory)
					r.Get("/get/{id}", userHandler.GetUserById)
					r.Get("/get", userHandler.GetAllUsers)
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

	go func() {
		generate := func() {
			if err := trainingService.GenerateOccurrences(context.Background()); err != nil {
				log.Printf("recurring trainings: generate occurrences: %v", err)
			}
		}

		generate()

		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			generate()
		}
	}()

	go func() {
		deactivate := func() {
			if err := trainingService.DeactivatePastTrainings(context.Background()); err != nil {
				log.Printf("trainings: deactivate past: %v", err)
			}
		}

		deactivate()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			deactivate()
		}
	}()

	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers,
		"OnCreateTraining", "OnUpdateTraining", "OnNewUser", "OnUserApprovalChanged", "OnTrainingPlaceAvailable")
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Printf("kafka producer: close: %v", err)
		}
	}()

	go func() {
		dispatch := func() {
			entries, err := outboxRepo.FindUnpublished(context.Background(), 100)
			if err != nil {
				log.Printf("outbox: find unpublished: %v", err)
				return
			}

			for _, entry := range entries {
				if err := kafkaProducer.PublishRaw(context.Background(), entry.Topic, entry.Payload); err != nil {
					log.Printf("outbox: publish entry %d to %q: %v", entry.Id, entry.Topic, err)
					continue
				}

				if err := outboxRepo.MarkPublished(context.Background(), entry.Id); err != nil {
					log.Printf("outbox: mark entry %d published: %v", entry.Id, err)
				}
			}
		}

		dispatch()

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			dispatch()
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
