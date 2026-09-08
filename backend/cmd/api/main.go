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
	trainingService := training.NewTrainingService(psqlTrainingRepo, outboxRepo)

	psqlBookingRepo := booking.NewPostgresBookingRepository(pool)
	bookingService := booking.NewBookingService(psqlBookingRepo, psqlTrainingRepo, psqlUserRepo, psqlIdentityRepo)
	bookingHandler := booking.NewBookingHandler(bookingService)

	psqlRecurringRepo := training.NewPostgresRecurringTrainingRepository(pool)
	recurringService := training.NewRecurringService(psqlRecurringRepo, psqlTrainingRepo)

	trainingHandler := training.NewTrainingHandler(trainingService, recurringService)

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
					r.Post("/update", trainingHandler.UpdateTraining)
					r.Post("/delete", trainingHandler.DeleteTraining)
					r.Post("/recurring/create", trainingHandler.CreateRecurringTraining)
					r.Get("/recurring/get", trainingHandler.GetRecurringTrainings)
					r.Post("/recurring/update", trainingHandler.UpdateRecurringTraining)
				})
				r.Get("/get/{id}", trainingHandler.GetTrainingById)
				r.Get("/get", trainingHandler.GetAllTrainings)
			})
			r.Route("/booking", func(r chi.Router) {
				r.Get("/getByUserId", bookingHandler.GetBookingsByUserId)
				r.Get("/getByTrainingId", bookingHandler.GetParticipants)
				r.Post("/signUp", bookingHandler.SignupForTraining)
				r.Post("/cancel", bookingHandler.CancelBooking)
				r.Group(func(r chi.Router) {
					r.Use(auth.RequireAdmin)
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

	// Регулярные тренировки материализуются, а не считаются на лету: каждое
	// сгенерированное занятие — обычная строка в trainings, на которую можно
	// записаться/отредактировать/отметить посещаемость как на любую другую.
	// Идемпотентно (ON CONFLICT DO NOTHING в репозитории), так что не страшно
	// звать чаще, чем реально нужно — лишний тик просто ничего не делает.
	go func() {
		generate := func() {
			if err := recurringService.GenerateOccurrences(context.Background()); err != nil {
				log.Printf("recurring trainings: generate occurrences: %v", err)
			}
		}

		generate() // сразу при старте — не ждать первого тика после деплоя

		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			generate()
		}
	}()

	// Прошедшие тренировки сами становятся неактивными — не показываются
	// обычным пользователям (см. TrainingService.GetAll), но остаются
	// доступны админу как история. Тикаем часто (не 6 часов, как у
	// регулярных тренировок) — иначе тренировка ещё долго после начала
	// выглядела бы как открытая для записи в списке. CreateBooking всё равно
	// не полагается только на этот тикер — проверяет дату сам, так что окно
	// между тиками ничем не рискует, только визуально.
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

	// Продьюсер теперь используется только релеем аутбокса ниже — сервисы
	// сами в Kafka больше не ходят, только пишут в outbox (см. platform/outbox).
	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers,
		"OnCreateTraining", "OnUpdateTraining", "OnNewUser")
	defer kafkaProducer.Close()

	// Outbox-релей: периодически вычитывает ещё не опубликованные записи
	// (см. platform/outbox) и шлёт их в Kafka как есть — payload там уже
	// сериализован сервисом, который его туда положил, релей его не трогает.
	// Тикаем часто — цель в том, чтобы задержка между "событие записано" и
	// "событие в Kafka" была небольшой, а не в том, чтобы экономить нагрузку.
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
