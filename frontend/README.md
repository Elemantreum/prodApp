# ProdApp frontend

React + TypeScript (Vite) mini app for the ProdApp booking backend. Works
both as a **Telegram Mini App** and a **VK Mini App** — platform is
auto-detected at startup, no build-time flag needed.

## How platform detection & auth works

- **Telegram**: `index.html` loads `telegram-web-app.js`, which injects
  `window.Telegram.WebApp`. If `WebApp.initData` is non-empty, we're in
  Telegram — that raw signed string is sent as `Authorization: tma <initData>`
  to `POST /api/v1/auth/telegram`.
- **VK**: VK always launches mini apps with `vk_user_id`, `sign`, etc.
  appended to the URL query string. If those are present, we're in VK — the
  raw query string is forwarded to `POST /api/v1/auth/vk?<query>` (via
  `@vkontakte/vk-bridge` for the rest of the platform integration: theme,
  `VKWebAppInit`).
- **Neither** (plain browser): shows a "open inside Telegram or VK" screen.
  In dev builds (`npm run dev`) this screen also has a manual token-paste
  box so you can develop/test without a real Telegram/VK client — get a
  token with `curl` against the backend's `/auth/telegram` or `/auth/vk`
  and paste it in.

The resulting JWT is cached in `localStorage`; on 401 the API client
silently re-runs the platform login and retries once (see
`src/api/client.ts` / `src/auth/AuthContext.tsx`).

## Layout

```
src/
  api/         fetch client + types mirroring the Go structs 1:1
  platform/    Telegram / VK SDK wrappers, theme mapping, platform detection
  auth/        AuthContext (login flow) + unverified JWT decode for UI state
  hooks/       useTelegramMainButton / useTelegramBackButton
  components/  presentational pieces (Button, cards, tab bar, ...)
  pages/       TrainingsPage, MyBookingsPage, AdminCreateTrainingPage
  utils/       date/duration formatting (backend sends duration in ns)
```

No VKUI / Telegram UI Kit — a single custom, minimal UI themed via CSS
variables fed from each platform's real theme (`--bg`, `--accent`, etc. in
`src/styles/global.css`), so it looks native-ish on both without maintaining
two component sets.

## Running

Local dev, hitting a backend on `localhost:8080`:

```bash
npm install
npm run dev
```

`vite.config.ts` sets `server.host: true` so you can tunnel the dev server
(ngrok/cloudflared) and open it for real inside Telegram/VK during
development.

Via docker-compose (serves the production build behind nginx, which proxies
`/api/*` and `/swagger/*` to the `backend` service — see `nginx.conf`):

```bash
docker compose up -d --build frontend
```

## Registering the mini app

- **Telegram**: `@BotFather` → your bot → *Bot Settings → Menu Button* (or
  *Mini App*), point it at your deployed frontend URL (must be HTTPS).
- **VK**: VK Apps platform → create a "Mini App", set the base URL to your
  deployed frontend. `TELEGRAM_BOT_TOKEN` / `VK_APP_SECRET_KEY` on the
  **backend** must match — that's what verifies the signed launch data (see
  `backend/internal/platform/auth/authenticator.go`).

## Known gaps

- No training edit/delete UI — the backend doesn't expose those routes yet
  (only `create`), so there's nothing to wire up.
- `GET /trainings/get` doesn't return how many spots are already taken, only
  `maxQuantity` — the training list can't show "3/10 занято" without a
  backend change to include booked counts.
