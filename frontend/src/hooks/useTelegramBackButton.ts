import { useEffect } from "react";
import { getTelegramWebApp } from "../platform/telegram";

/** Shows Telegram's native back button and wires it to `onBack` while the
 *  owning page is mounted. No-ops on VK/browser. */
export function useTelegramBackButton(onBack: (() => void) | null): void {
  useEffect(() => {
    const tg = getTelegramWebApp();
    const button = tg?.BackButton;
    if (!button || !onBack) return undefined;

    button.onClick(onBack);
    button.show();

    return () => {
      button.offClick(onBack);
      button.hide();
    };
  }, [onBack]);
}
