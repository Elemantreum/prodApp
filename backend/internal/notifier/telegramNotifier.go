package notifier

import (
	"context"
	"log"
	"prodApp/internal/user"

	"github.com/mr-linch/go-tg"
)

type TelegramNotifier struct {
	client      *tg.Client
	userService *user.UserService
}

func NewTelegramNotifier(telegramBotToken string, userService *user.UserService) *TelegramNotifier {
	return &TelegramNotifier{
		client:      tg.New(telegramBotToken),
		userService: userService,
	}
}

func (t *TelegramNotifier) NotifyMany(ctx context.Context, chatId []int, msg string) error {
	for _, id := range chatId {
		_, err := t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)

		log.Default().Println(err)
	}

	return nil
}

func (t *TelegramNotifier) Notify(ctx context.Context, chatId int, msg string) error {
	_, err := t.client.SendMessage(tg.ChatID(chatId), msg).Do(ctx)

	return err
}

//
//func (t TelegramNotifier) NotifyAll(ctx context.Context, msg string) error {
//	userList, err := t.userService.ListUsers(ctx)
//	if err != nil {
//		return err
//	}
//	onlyUserIds := make([]int, 0)
//	for _, user := range userList {
//		onlyUserIds = append(onlyUserIds, user.Id)
//	}
//	tgUserList, err := t.userService.GetBatchTgById(ctx, onlyUserIds, "telegram")
//
//	for _, user := range tgUserList {
//		id, err := strconv.ParseInt(user.ExternalId, 10, 64)
//		if err != nil {
//			log.Default().Println(err)
//			continue
//		}
//		t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
//	}
//	return err
//}
//
//func (t *TelegramNotifier) NotifyAdmins(ctx context.Context, msg string) error {
//	adminList, err := t.userService.ListAdmins(ctx)
//	onlyAdminIds := make([]int, 0)
//	for _, admin := range adminList {
//		onlyAdminIds = append(onlyAdminIds, admin.Id)
//	}
//
//	if err != nil {
//		return err
//	}
//	tgAdminList, err := t.userService.GetBatchTgById(ctx, onlyAdminIds, "telegram")
//	if err != nil {
//		return err
//	}
//
//	for _, admin := range tgAdminList {
//		id, err := strconv.ParseInt(admin.ExternalId, 10, 64)
//		if err != nil {
//			return err
//		}
//		t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
//	}
//	return nil
//}
