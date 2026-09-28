import bridge from "@vkontakte/vk-bridge";

export function isVk(): boolean {
  const params = new URLSearchParams(window.location.search);
  return params.has("vk_user_id") && params.has("sign");
}

const capturedLaunchQuery = window.location.search.replace(/^\?/, "");

export function getVkLaunchQuery(): string {
  return capturedLaunchQuery;
}

export async function initVk(): Promise<void> {
  await bridge.send("VKWebAppInit", {});
}

export type VkScheme = "light" | "dark";

export function subscribeVkTheme(onChange: (scheme: VkScheme) => void): () => void {
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
