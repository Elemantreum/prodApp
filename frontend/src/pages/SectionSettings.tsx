import { useState, type FormEvent } from "react";
import { useSection } from "../auth/Sections";
import { api } from "../api/client";
import type { Section } from "../api/types";
import { AdminDialog, AdminSearch, useAdminCreate } from "../components/AdminUI";
import { ErrorBanner } from "../components/ErrorBanner";
import { t } from "../i18n";

function SectionForm({ section, onSaved }: { section?: Section; onSaved?: () => void }) {
  const { reload } = useSection();
  const [name, setName] = useState(section?.name ?? "");
  const [logo, setLogo] = useState(section?.logo ?? "");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  async function save(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api.saveSection(section?.id ?? 0, name.trim(), logo);
      await reload();
      setMessage(t("Секция сохранена"));
      onSaved?.();
    } catch (err) {
      setMessage(String(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <form onSubmit={save}>
      <label>
        {t("Название секции")}
        <input required maxLength={120} value={name} onChange={(e) => setName(e.target.value)} />
      </label>
      <label>
        {t("Ссылка на логотип")}
        <input
          type="url"
          value={logo.startsWith("data:") ? "" : logo}
          onChange={(e) => setLogo(e.target.value)}
        />
      </label>
      {logo && <img src={logo} alt={t("Логотип секции")} width={40} height={40} />}
      <label>
        {t("Файл логотипа (PNG или JPEG, до 256 КБ)")}
        <input
          type="file"
          accept="image/png,image/jpeg"
          disabled={busy}
          onChange={(e) => {
            const file = e.target.files?.[0];
            if (!file) return;
            if (file.size > 256 * 1024) {
              setMessage(t("Файл слишком большой"));
              return;
            }
            const reader = new FileReader();
            setBusy(true);
            reader.onload = () => {
              if (typeof reader.result === "string") setLogo(reader.result);
            };
            reader.onerror = () => setMessage(t("Не удалось прочитать файл"));
            reader.onloadend = () => setBusy(false);
            reader.readAsDataURL(file);
          }}
        />
      </label>
      <div>
        <button className="btn primary" disabled={busy || !name.trim()}>
          {t("Сохранить")}
        </button>
      </div>
      {message && <p role="status">{message}</p>}
    </form>
  );
}

export function SectionSettings() {
  const { current } = useSection();
  return (
    <>
      <section className="card">
        <h2>{t("Оформление")}</h2>
        <p className="muted">{t("Название и логотип видны всем участникам секции.")}</p>
        <SectionForm key={current.id} section={current} />
      </section>
      <SectionVenues />
    </>
  );
}

function SectionVenues() {
  const { current, reload } = useSection();
  const [name, setName] = useState("");
  const [confirm, setConfirm] = useState<string>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const venues = current.venues ?? [];

  async function change(action: () => Promise<unknown>) {
    setBusy(true);
    setError("");
    try {
      await action();
      await reload();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("Не удалось сохранить"));
    } finally {
      setBusy(false);
    }
  }

  function add(event: FormEvent) {
    event.preventDefault();
    setConfirm(undefined);
    void change(async () => {
      await api.addVenue(current.id, name.trim());
      setName("");
    });
  }

  function remove(venue: string) {
    void change(async () => {
      await api.deleteVenue(current.id, venue);
      setConfirm(undefined);
    });
  }

  return (
    <section className="card" aria-labelledby="section-venues-heading">
      <h2 id="section-venues-heading">{t("Места проведения")}</h2>
      <form className="venue-form" onSubmit={add}>
        <label>
          {t("Название места")}
          <input
            maxLength={120}
            required
            disabled={busy}
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </label>
        <button type="submit" className="btn primary" disabled={busy || !name.trim()}>
          {t("Добавить")}
        </button>
      </form>
      {error && <ErrorBanner message={error} />}
      {venues.length === 0 && <p className="muted">{t("Мест проведения пока нет.")}</p>}
      <ul>
        {venues.map((venue) => (
          <li className="row" key={venue}>
            <span className="grow">{venue}</span>
            {confirm === venue ? (
              <>
                <button type="button" className="btn" disabled={busy} onClick={() => setConfirm(undefined)}>
                  {t("Отмена")}
                </button>
                <button type="button" className="btn danger" disabled={busy} onClick={() => remove(venue)}>
                  {t("Подтвердить")}
                </button>
              </>
            ) : (
              <button type="button" className="btn" disabled={busy} onClick={() => setConfirm(venue)}>
                {t("Удалить")}
              </button>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function AdminSectionsPage() {
  const { directory, reload } = useSection();
  const [create, setCreate] = useAdminCreate();
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<Section>();
  const [confirm, setConfirm] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  function close() {
    setCreate(false);
    setSelected(undefined);
    setConfirm(false);
    setError("");
  }
  const sections = directory.sections.filter((s) => s.name.toLowerCase().includes(query.trim().toLowerCase()));
  return (
    <section className="card">
      <AdminSearch value={query} onChange={setQuery} />
      {sections.map((s) => (
        <div className="row" key={s.id}>
          <span className="avatar">{s.logo ? <img src={s.logo} alt="" /> : s.name[0]}</span>
          <div className="grow">
            <b>{s.name}</b>
            <small>{t("Действующая секция")}</small>
          </div>
          <button className="btn" onClick={() => setSelected(s)}>
            {t("Управление")}
          </button>
        </div>
      ))}
      {!sections.length && <p>{t("Ничего не найдено")}</p>}
      {(create || selected) && (
        <AdminDialog title={selected?.name ?? t("Создать секцию")} onClose={close} busy={busy}>
          <SectionForm section={selected} onSaved={close} />
          {selected && selected.id !== 1 && (
            <button
              className="btn danger"
              disabled={busy}
              onClick={async () => {
                if (!confirm) {
                  setConfirm(true);
                  return;
                }
                setBusy(true);
                try {
                  await api.deleteSection(selected.id);
                  await reload();
                  close();
                } catch (e) {
                  setError(String(e));
                } finally {
                  setBusy(false);
                }
              }}
            >
              {confirm ? t("Подтвердить удаление секции") : t("Удалить секцию")}
            </button>
          )}
          {error && <p role="alert">{error}</p>}
        </AdminDialog>
      )}
    </section>
  );
}
