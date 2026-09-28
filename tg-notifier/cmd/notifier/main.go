package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"prodApp/notifier-service/internal/config"
	"prodApp/notifier-service/internal/events"
	"prodApp/notifier-service/internal/kafka"
	"prodApp/notifier-service/internal/notifier"
)

var moscowTZ = time.FixedZone("MSK", 3*60*60)

func main() {
	log.Printf("tg-notifier: start")

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tgNotifier := notifier.NewTelegramNotifier(cfg.TelegramBotToken)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup,
			func(ctx context.Context, e events.OnCreateTraining) {
				text := fmt.Sprintf("Новая тренировка! %s в %s", e.Description, e.Date.In(moscowTZ).Format("15:04 02.01"))
				if err := tgNotifier.NotifyMany(ctx, e.ChatIds, text); err != nil {
					log.Printf("training created: notify: %v", err)
				}
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup,
			func(ctx context.Context, e events.OnUpdateTraining) {
				text := formatTrainingUpdate(e)
				if err := tgNotifier.NotifyMany(ctx, e.ChatIds, text); err != nil {
					log.Printf("training updated: notify: %v", err)
				}
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup,
			func(ctx context.Context, e events.OnNewUser) {
				name := e.FirstName
				if e.LastName != "" {
					name += " " + e.LastName
				}
				text := fmt.Sprintf("Новый пользователь: %s (%s)", name, e.Provider)
				if err := tgNotifier.NotifyMany(ctx, e.ChatIds, text); err != nil {
					log.Printf("user created: notify: %v", err)
				}
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Consume(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup,
			func(ctx context.Context, e events.OnUserApprovalChanged) {
				text := "🎉 Тебя одобрили! Теперь можно записываться на тренировки."
				if !e.Approved {
					text = "⚠️ С тебя сняли одобрение — записываться на тренировки временно нельзя."
				}
				if err := tgNotifier.Notify(ctx, e.ChatId, text); err != nil {
					log.Printf("user approval changed: notify: %v", err)
				}
			})
	}()

	wg.Wait()
	log.Printf("tg-notifier: shutting down")
}

func formatTrainingUpdate(e events.OnUpdateTraining) string {
	var changes []string

	if e.OldDescription != e.Description {
		changes = append(changes, fmt.Sprintf("описание: «%s» → «%s»", e.OldDescription, e.Description))
	}
	if !e.OldDate.Equal(e.Date) {
		changes = append(changes, fmt.Sprintf("время: %s → %s",
			e.OldDate.In(moscowTZ).Format("15:04 02.01"), e.Date.In(moscowTZ).Format("15:04 02.01")))
	}
	if e.OldMaxQuantity != e.MaxQuantity {
		changes = append(changes, fmt.Sprintf("мест: %d → %d", e.OldMaxQuantity, e.MaxQuantity))
	}
	if e.OldDuration != e.Duration {
		changes = append(changes, fmt.Sprintf("длительность: %d мин → %d мин",
			int(e.OldDuration.Minutes()), int(e.Duration.Minutes())))
	}
	if e.OldIsActive != e.IsActive {
		status := map[bool]string{true: "активна", false: "неактивна"}
		changes = append(changes, fmt.Sprintf("статус: %s → %s", status[e.OldIsActive], status[e.IsActive]))
	}

	if len(changes) == 0 {
		return fmt.Sprintf("Тренировка обновлена: %s", e.Description)
	}

	return fmt.Sprintf("Тренировка изменена: %s\n%s", e.Description, strings.Join(changes, "\n"))
}
