import { useEffect, useId, useRef, type ReactNode } from "react";
import { useSearchParams } from "react-router-dom";
import { t } from "../i18n";

export function useAdminCreate() {
  const [params, setParams] = useSearchParams();
  function setOpen(open: boolean) {
    setParams(
      (previous) => {
        const next = new URLSearchParams(previous);
        if (open) next.set("new", "1");
        else next.delete("new");
        return next;
      },
      { replace: true },
    );
  }
  return [params.get("new") === "1", setOpen] as const;
}

export function AdminDialog({
  title,
  onClose,
  children,
  busy,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
  busy?: boolean;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => {
    const dialog = ref.current!;
    dialog.showModal();
    return () => dialog.close();
  }, []);
  return (
    <dialog
      ref={ref}
      className="admin-dialog"
      aria-labelledby={id}
      onCancel={(e) => {
        e.preventDefault();
        if (!busy) onClose();
      }}
    >
      <div className="admin-dialog-head">
        <h2 id={id}>{title}</h2>
        <button className="btn" disabled={busy} onClick={onClose}>
          {t("Закрыть")}
        </button>
      </div>
      {children}
    </dialog>
  );
}

export function AdminSearch({ value, onChange }: { value: string; onChange: (s: string) => void }) {
  return (
    <input
      className="field"
      type="search"
      aria-label={t("Поиск")}
      placeholder={t("Поиск по названию или имени")}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}
