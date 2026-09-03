import { useEffect } from "react";
import { getTelegramWebApp } from "../platform/telegram";

interface Options {
  text: string;
  onClick: () => void;
  visible?: boolean;
  disabled?: boolean;
  loading?: boolean;
}

/** Maps a page action onto Telegram's native bottom MainButton when running
 *  inside Telegram. No-ops harmlessly on VK/browser. */
export function useTelegramMainButton({ text, onClick, visible = true, disabled = false, loading = false }: Options): void {
  useEffect(() => {
    const tg = getTelegramWebApp();
    const button = tg?.MainButton;
    if (!button) return undefined;

    button.setText(text);
    button.onClick(onClick);

    if (visible) button.show();
    else button.hide();

    if (disabled) button.disable();
    else button.enable();

    if (loading) button.showProgress(true);
    else button.hideProgress();

    return () => {
      button.offClick(onClick);
      button.hide();
    };
  }, [text, onClick, visible, disabled, loading]);
}
