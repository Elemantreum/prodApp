import { useCallback, useEffect, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import { api, ApiError } from "../api/client";
import { Button } from "../components/Button";
import { ErrorBanner } from "../components/ErrorBanner";
import { Spinner } from "../components/Spinner";

type Status = "checking" | "needs-registration" | "ready";

/**
 * Обязательная "регистрация" при первом входе: пока User.isRegistered
 * false, вместо остального приложения показываем один единственный экран —
 * "как вас зовут" — и не пускаем дальше, пока не сохранят. Флаг переводится
 * в true на бэке первым же успешным UpdateProfile (см. UserService), так
 * что сам сабмит — это обычный updateProfile, никакой отдельной ручки не
 * нужно. Сидит внутри LoginGate: к этому моменту токен уже есть.
 */
export function RegistrationGate({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<Status>("checking");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const check = useCallback(async () => {
    setStatus("checking");
    setError(null);
    try {
      const profile = await api.getOwnProfile();
      if (profile.isRegistered) {
        setStatus("ready");
      } else {
        setFirstName(profile.firstName);
        setLastName(profile.lastName);
        setStatus("needs-registration");
      }
    } catch (e) {
      // Не блокируем намертво, если сам проверочный запрос упал — пусть
      // остальное приложение само разбирается с ошибкой при следующем вызове.
      setStatus("ready");
      void e;
    }
  }, []);

  useEffect(() => {
    check();
  }, [check]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await api.updateProfile(firstName.trim(), lastName.trim());
      setStatus("ready");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Не удалось сохранить");
    } finally {
      setSubmitting(false);
    }
  }

  if (status === "checking") {
    return (
      <div className="gate">
        <Spinner size={32} />
      </div>
    );
  }

  if (status === "needs-registration") {
    return (
      <div className="gate">
        <div className="gate__form">
          <h1 className="gate__title">Добро пожаловать</h1>
          <p>Прежде чем продолжить, представьтесь — это займёт секунду.</p>
          <form className="form" onSubmit={handleSubmit}>
            <label className="form__field">
              <span>Имя</span>
              <input type="text" value={firstName} onChange={(e) => setFirstName(e.target.value)} required autoFocus />
            </label>
            <label className="form__field">
              <span>Фамилия</span>
              <input type="text" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </label>
            {error && <ErrorBanner message={error} />}
            <Button type="submit" loading={submitting} disabled={!firstName.trim()}>
              Продолжить
            </Button>
          </form>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
