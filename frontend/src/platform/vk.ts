import bridge from "@vkontakte/vk-bridge";

/** VK always launches a mini app with these appended to the URL query string. */
export function isVk(): boolean {
  const params = new URLSearchParams(window.location.search);
  return params.has("vk_user_id") && params.has("sign");
}

/** The raw launch-params query string, exactly as VK gave it — this is what
 *  the backend's HMAC signature check (auth/authenticator.go) verifies. */
export function getVkLaunchQuery(): string {
  return window.location.search.replace(/^\?/, "");
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
