import { useState } from "react";
import type { ReactNode } from "react";
import { useAuth } from "../auth/AuthContext";
import { Spinner } from "../components/Spinner";
import { Button } from "../components/Button";

export function LoginGate({ children }: { children: ReactNode }) {
  const { status, error, retry, platform, loginWithToken } = useAuth();

  if (status === "authenticated") return <>{children}</>;

  if (status === "loading") {
    return (
      <div className="gate">
        <Spinner size={32} />
        <p>Входим…</p>
      </div>
    );
  }

  if (status === "error") {
    return (
      <div className="gate">
        <p className="gate__error">{error}</p>
        <Button onClick={retry}>Повторить</Button>
      </div>
    );
  }

  return (
    <div className="gate">
      <p>Открой это приложение внутри Telegram или VK.</p>
      {import.meta.env.DEV && <DevTokenPanel platform={platform} onSubmit={loginWithToken} />}
    </div>
  );
}

function DevTokenPanel({ platform, onSubmit }: { platform: string; onSubmit: (token: string) => void }) {
  const [value, setValue] = useState("");

  return (
    <div className="dev-panel">
      <p className="dev-panel__hint">
        Dev-режим ({platform}): вставь access_token, полученный вручную от{" "}
        <code>POST /api/v1/auth/telegram</code> или <code>/auth/vk</code>, чтобы посмотреть приложение без
        реального Telegram/VK.
      </p>
      <textarea value={value} onChange={(e) => setValue(e.target.value)} rows={3} placeholder="eyJhbGciOi..." />
      <Button disabled={!value.trim()} onClick={() => onSubmit(value.trim())}>
        Войти с токеном
      </Button>
    </div>
  );
}
