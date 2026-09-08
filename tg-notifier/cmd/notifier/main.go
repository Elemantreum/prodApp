package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"prodApp/notifier-service/internal/config"
	"prodApp/notifier-service/internal/events"
	"prodApp/notifier-service/internal/kafka"
	"prodApp/notifier-service/internal/notifier"
	"prodApp/notifier-service/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Printf("tg-notifier: start")

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	recipients := repository.NewRecipients(pool)
	tgNotifier := notifier.NewTelegramNotifier(cfg.TelegramBotToken)

	var wg sync.WaitGroup

	// Каждое событие — свой топик, свой consumer-луп, своя интерпретация
	// "кого уведомлять и что написать". Транспортная часть (чтение,
	// JSON) — общая (см. kafka.Consume), бизнес-логика — только тут.
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, "OnCreateTraining",
			func(ctx context.Context, e events.OnCreateTraining) {
				chatIDs, err := recipients.AllTelegramChatIDs(ctx)
				if err != nil {
					log.Printf("training created: resolve recipients: %v", err)
					return
				}
				text := fmt.Sprintf("Новая тренировка! %s в %s", e.Description, e.Date.Format("15:04 02.01"))
				if err := tgNotifier.NotifyMany(ctx, chatIDs, text); err != nil {
					log.Printf("training created: notify: %v", err)
				}
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, "OnUpdateTraining",
			func(ctx context.Context, e events.OnUpdateTraining) {
				chatIDs, err := recipients.AllTelegramChatIDs(ctx)
				if err != nil {
					log.Printf("training updated: resolve recipients: %v", err)
					return
				}
				text := fmt.Sprintf("Тренировка изменена: %s, теперь %s", e.Description, e.Date.Format("15:04 02.01"))
				if err := tgNotifier.NotifyMany(ctx, chatIDs, text); err != nil {
					log.Printf("training updated: notify: %v", err)
				}
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, "OnNewUser",
			func(ctx context.Context, e events.OnNewUser) {
				chatIDs, err := recipients.AdminTelegramChatIDs(ctx)
				if err != nil {
					log.Printf("user created: resolve recipients: %v", err)
					return
				}
				name := e.FirstName
				if e.LastName != "" {
					name += " " + e.LastName
				}
				text := fmt.Sprintf("Новый пользователь: %s (%s)", name, e.Provider)
				if err := tgNotifier.NotifyMany(ctx, chatIDs, text); err != nil {
					log.Printf("user created: notify: %v", err)
				}
			})
	}()

	wg.Wait()
	log.Printf("tg-notifier: shutting down")
}
