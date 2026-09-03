import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Telegram/VK mini apps are opened inside a webview through the platform's
// own domain (t.me / vk.com iframe) — the app itself is always served from
// its own origin behind HTTPS in production. In dev, `host: true` lets you
// open the Vite dev server through a tunnel (ngrok/cloudflared) so it can
// actually be tested inside a real Telegram/VK client.
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
  },
  preview: {
    host: true,
    port: 4173,
  },
});
