package notifier

import (
	"context"
	"prodApp/internal/booking"
	"prodApp/internal/user"
	"strconv"

	"github.com/mr-linch/go-tg"
)

type TelegramNotifier struct {
	client         *tg.Client
	userService    *user.UserService
	bookingService *booking.BookingService
}

func NewTelegramNotifier(telegramBotToken string, userService *user.UserService, bookingService *booking.BookingService) *TelegramNotifier {
	return &TelegramNotifier{
		client:         tg.New(telegramBotToken),
		userService:    userService,
		bookingService: bookingService,
	}
}

func (t *TelegramNotifier) NotifyUser(ctx context.Context, chatID int64, msg string) error {
	_, err := t.client.SendMessage(tg.ChatID(chatID), msg).Do(ctx)
	return err
}

func (t TelegramNotifier) notifyAll(ctx context.Context, msg string) error {
	userList, err := t.userService.ListUsers(ctx)
	if err != nil {
		return err
	}
	onlyUserIds := make([]int, 0)
	for _, user := range userList {
		onlyUserIds = append(onlyUserIds, user.Id)
	}
	tgUserList, err := t.userService.GetBatchTgById(ctx, onlyUserIds, "telegram")

	for _, user := range tgUserList {
		id, err := strconv.ParseInt(user.ExternalId, 10, 64)
		if err != nil {
			return err
		}
		t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
	}
	return nil
}

func (t *TelegramNotifier) NotifyAdmins(ctx context.Context, msg string) error {
	adminList, err := t.userService.ListAdmins(ctx)
	onlyAdminIds := make([]int, 0)
	for _, admin := range adminList {
		onlyAdminIds = append(onlyAdminIds, admin.Id)
	}

	if err != nil {
		return err
	}
	tgAdminList, err := t.userService.GetBatchTgById(ctx, onlyAdminIds, "telegram")
	if err != nil {
		return err
	}

	for _, admin := range tgAdminList {
		id, err := strconv.ParseInt(admin.ExternalId, 10, 64)
		if err != nil {
			return err
		}
		t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
	}
	return nil
}

func (t *TelegramNotifier) NotifyTraining(ctx context.Context, trainingId int, msg string) error {
	userList, err := t.bookingService.GetParticipants(ctx, trainingId)
	if err != nil {
		return err
	}

	onlyUserIds := make([]int, 0)
	for _, user := range userList {
		onlyUserIds = append(onlyUserIds, user.UserId)
	}
	tgUserList, err := t.userService.GetBatchTgById(ctx, onlyUserIds, "telegram")

	for _, user := range tgUserList {
		id, err := strconv.ParseInt(user.ExternalId, 10, 64)
		if err != nil {
			return err
		}
		t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
	}
	return nil
}
