package sectioninvite

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	http2 "prodApp/internal/platform/http"
	"prodApp/internal/section"
)

func TestInvitationLifecycle(t *testing.T) {
	pool := invitationTestDB(t)
	handler := NewInvitationHandler(NewInvitationService(NewPostgresInvitationRepository(pool)))
	router := chi.NewRouter()
	router.Get("/sections/invites", handler.List)
	router.Post("/sections/invites", handler.Create)
	router.Delete("/sections/invites/{id}", handler.Revoke)
	router.Post("/sections/invites/preview", handler.Preview)
	router.Post("/sections/invites/redeem", handler.Redeem)

	request := func(method, path, body string, userID, sectionID int) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/sections/invites"+path, strings.NewReader(body))
		ctx := http2.ContextWithClaims(section.WithID(r.Context(), sectionID), &http2.Claims{UserID: userID})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r.WithContext(ctx))
		return w
	}

	if w := request(http.MethodGet, "", "", 1, 2); w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("empty list: %d %s", w.Code, w.Body)
	}
	for _, body := range []string{
		`{"expiresAt":"2099-01-02T03:04:05+03:00","maxUses":0}`,
		`{"expiresAt":"2000-01-01T00:00:00Z","maxUses":1}`,
		`{"expiresAt":"tomorrow","maxUses":1}`,
	} {
		if w := request(http.MethodPost, "", body, 1, 2); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid create %s: %d %s", body, w.Code, w.Body)
		}
	}
	w := request(http.MethodPost, "", `{"expiresAt":"2099-01-02T03:04:05+03:00","maxUses":1}`, 1, 2)
	var invite struct {
		ID        int    `json:"id"`
		Code      string `json:"code"`
		ExpiresAt string `json:"expiresAt"`
	}
	if w.Code != http.StatusCreated || json.Unmarshal(w.Body.Bytes(), &invite) != nil || invite.ExpiresAt != "2099-01-02T00:04:05Z" {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	if raw, err := base64.RawURLEncoding.DecodeString(invite.Code); err != nil || len(raw) < 16 {
		t.Fatalf("invalid base64url code, length %d: %v", len(raw), err)
	}
	var hash string
	digest := sha256.Sum256([]byte(invite.Code))
	if err := pool.QueryRow(t.Context(), "SELECT token_hash FROM section_invites WHERE id=$1", invite.ID).Scan(&hash); err != nil || hash != hex.EncodeToString(digest[:]) {
		t.Fatalf("stored token is not its SHA-256 digest: %v", err)
	}
	if w := request(http.MethodGet, "", "", 1, 2); w.Code != http.StatusOK || strings.Contains(w.Body.String(), `"code"`) {
		t.Fatalf("list exposed code or failed: %d %s", w.Code, w.Body)
	}
	path := fmt.Sprintf("/%d", invite.ID)
	if w := request(http.MethodDelete, path, "", 1, 3); w.Code != http.StatusNotFound {
		t.Fatalf("cross-section revoke: %d %s", w.Code, w.Body)
	}
	body := fmt.Sprintf(`{"code":%q}`, invite.Code)
	if w := request(http.MethodPost, "/preview", body, 5, 3); w.Code != http.StatusOK {
		t.Fatalf("preview: %d %s", w.Code, w.Body)
	}
	if w := request(http.MethodPost, "/redeem", body, 5, 3); w.Code != http.StatusForbidden {
		t.Fatalf("unregistered redeem: %d %s", w.Code, w.Body)
	}
	if _, err := pool.Exec(t.Context(), `UPDATE section_invites SET expires_at=now() WHERE id=$1`, invite.ID); err != nil {
		t.Fatal(err)
	}
	if w := request(http.MethodPost, "/redeem", body, 2, 3); w.Code != http.StatusConflict {
		t.Fatalf("expired redeem: %d %s", w.Code, w.Body)
	}
	if _, err := pool.Exec(t.Context(), `UPDATE section_invites SET expires_at=$2::text::timestamptz WHERE id=$1`, invite.ID, invite.ExpiresAt); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan struct{ userID, status int }, 2)
	for _, userID := range []int{2, 3} {
		go func() {
			<-start
			w := request(http.MethodPost, "/redeem", body, userID, 3)
			results <- struct{ userID, status int }{userID, w.Code}
		}()
	}
	close(start)
	first, second := <-results, <-results
	if first.status == http.StatusConflict {
		first, second = second, first
	}
	if first.status != http.StatusOK || second.status != http.StatusConflict {
		t.Fatalf("concurrent last use returned %d and %d", first.status, second.status)
	}
	if w := request(http.MethodPost, "/redeem", body, first.userID, 3); w.Code != http.StatusOK {
		t.Fatalf("repeat redeem at capacity: %d %s", w.Code, w.Body)
	}
	if _, err := pool.Exec(t.Context(), `UPDATE section_invites SET max_uses=2 WHERE id=$1`, invite.ID); err != nil {
		t.Fatal(err)
	}
	if w := request(http.MethodDelete, path, "", 1, 2); w.Code != http.StatusNoContent {
		t.Fatalf("revoke: %d %s", w.Code, w.Body)
	}
	if w := request(http.MethodPost, "/redeem", body, 1, 3); w.Code != http.StatusOK {
		t.Fatalf("existing admin redeem after revocation: %d %s", w.Code, w.Body)
	}
	if w := request(http.MethodPost, "/preview", body, second.userID, 3); w.Code != http.StatusConflict {
		t.Fatalf("revoked preview: %d %s", w.Code, w.Body)
	}
	if w := request(http.MethodPost, "/redeem", body, second.userID, 3); w.Code != http.StatusConflict {
		t.Fatalf("revoked redeem: %d %s", w.Code, w.Body)
	}

	var uses int
	var admin, profileLocked, originalAdmin, loserJoined bool
	err := pool.QueryRow(t.Context(), `
		SELECT i.uses, m.is_admin, u.profile_locked,
		       (SELECT is_admin FROM section_members WHERE section_id=i.section_id AND user_id=1),
		       EXISTS (SELECT 1 FROM section_members WHERE section_id=i.section_id AND user_id=$3)
		FROM section_invites i JOIN section_members m ON m.section_id=i.section_id
		JOIN users u ON u.id=m.user_id WHERE i.id=$1 AND m.user_id=$2`, invite.ID, first.userID, second.userID).
		Scan(&uses, &admin, &profileLocked, &originalAdmin, &loserJoined)
	if err != nil || uses != 1 || admin || !profileLocked || !originalAdmin || loserJoined {
		t.Fatalf("final state: uses=%d new admin=%t profile locked=%t original admin=%t loser joined=%t: %v", uses, admin, profileLocked, originalAdmin, loserJoined, err)
	}
}

func invitationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("SECTIONINVITE_TEST_DSN")
	if dsn == "" {
		t.Skip("SECTIONINVITE_TEST_DSN is unset")
	}
	ctx := t.Context()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("sectioninvite_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, `
		CREATE TABLE sections (id int PRIMARY KEY, name text NOT NULL, logo_url text NOT NULL);
		CREATE TABLE users (id int PRIMARY KEY, is_registered bool NOT NULL, profile_locked bool NOT NULL DEFAULT false);
		CREATE TABLE section_members (
			section_id int REFERENCES sections ON DELETE CASCADE, user_id int REFERENCES users ON DELETE CASCADE,
			is_admin bool NOT NULL, PRIMARY KEY (section_id, user_id)
		);
		CREATE TABLE section_invites (
			id serial PRIMARY KEY, section_id int NOT NULL REFERENCES sections ON DELETE CASCADE,
			token_hash text NOT NULL UNIQUE, created_by int REFERENCES users ON DELETE SET NULL,
			expires_at timestamptz NOT NULL, max_uses int NOT NULL, uses int NOT NULL DEFAULT 0,
			revoked_at timestamptz
		);
		CREATE FUNCTION lock_profile() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			UPDATE users SET profile_locked=true WHERE id=NEW.user_id;
			RETURN NEW;
		END;
		$$;
		CREATE TRIGGER lock_profile AFTER INSERT ON section_members FOR EACH ROW EXECUTE FUNCTION lock_profile();
		INSERT INTO sections VALUES (2, 'Target', 'logo'), (3, 'Other', '');
		INSERT INTO users (id, is_registered) VALUES (1,true), (2,true), (3,true), (5,false);
		INSERT INTO section_members VALUES (2,1,true);
	`); err != nil {
		t.Fatal(err)
	}
	return pool
}
