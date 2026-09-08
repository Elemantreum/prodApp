// Package repository — read-only доступ к той же Postgres, что и монолит.
// Никакого общего Go-кода с бэкендом (это отдельный модуль) — только два
// узких запроса, ровно то, что нужно этому сервису для решения "кому
// слать". Полную доменную модель (User, Identity и т.п.) сюда тащить незачем.
package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Recipients struct {
	pool *pgxpool.Pool
}

func NewRecipients(pool *pgxpool.Pool) *Recipients {
	return &Recipients{pool: pool}
}

// AllTelegramChatIDs — у кого вообще привязан Telegram, независимо от
// approve/admin статуса. Используется для широковещательных уведомлений
// (например "создана новая тренировка").
func (r *Recipients) AllTelegramChatIDs(ctx context.Context) ([]int, error) {
	return r.queryChatIDs(ctx, `SELECT external_id FROM user_identities WHERE provider = 'telegram'`)
}

// AdminTelegramChatIDs — только те, у кого is_admin = true. Используется
// для уведомлений, которые касаются только администраторов (например "зарегистрировался новый пользователь").
func (r *Recipients) AdminTelegramChatIDs(ctx context.Context) ([]int, error) {
	return r.queryChatIDs(ctx, `
		SELECT ui.external_id
		FROM user_identities ui
		JOIN users u ON u.id = ui.user_id
		WHERE ui.provider = 'telegram' AND u.is_admin = true`)
}

func (r *Recipients) queryChatIDs(ctx context.Context, sql string) ([]int, error) {
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query chat ids: %w", err)
	}
	defer rows.Close()

	chatIDs := make([]int, 0)
	for rows.Next() {
		var externalID string
		if err := rows.Scan(&externalID); err != nil {
			return nil, fmt.Errorf("scan chat id: %w", err)
		}

		id, err := strconv.Atoi(externalID)
		if err != nil {
			// Не должно случаться для telegram (external_id там всегда
			// числовой), но лучше пропустить кривую запись, чем уронить всю
			// рассылку остальным.
			continue
		}
		chatIDs = append(chatIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat ids: %w", err)
	}

	return chatIDs, nil
}
