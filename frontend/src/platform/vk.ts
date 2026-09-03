import bridge from "@vkontakte/vk-bridge";

/** VK always launches a mini app with these appended to the URL query string. */
export function isVk(): boolean {
  const params = new URLSearchParams(window.location.search);
  return params.has("vk_user_id") && params.has("sign");
}

/**
 * The raw launch-params query string, exactly as VK gave it — this is what
 * the backend's HMAC signature check (auth/authenticator.go) verifies.
 *
 * Captured once, at MODULE LOAD time (not lazily on first call). React
 * Router's client-side navigation (pushState to a plain path like
 * "/account") drops the query string from the URL — read this later (e.g.
 * from the account-linking screen) and window.location.search would already
 * be empty, silently turning the request into an unsigned one (backend
 * correctly rejects those with 401). A lazy "cache on first call" isn't
 * enough by itself either: if the user already has a valid cached token,
 * AuthContext's bootstrap skips the login call entirely, so nothing would
 * have called this before the user navigates away and empties the URL. This
 * module's top-level code runs once, synchronously, before React even
 * mounts — always before any navigation — so reading it here unconditionally
 * is the one point that's guaranteed safe.
 */
const capturedLaunchQuery = window.location.search.replace(/^\?/, "");

export function getVkLaunchQuery(): string {
  return capturedLaunchQuery;
}

export async function initVk(): Promise<void> {
  await bridge.send("VKWebAppInit", {});
}

export type VkScheme = "light" | "dark";

export function subscribeVkTheme(onChange: (scheme: VkScheme) => void): () => void {
  // vk-bridge's event typing is a large generated union; `any` here is
  // deliberate rather than fighting it for one narrow field lookup.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handler = (e: any) => {
    if (e?.detail?.type === "VKWebAppUpdateConfig") {
      const scheme: string | undefined = e.detail.data?.scheme;
      onChange(scheme?.includes("dark") || scheme?.includes("space") ? "dark" : "light");
    }
  };
  bridge.subscribe(handler);
  return () => bridge.unsubscribe(handler);
}
