import { Outlet, useLocation, useSearchParams } from "react-router-dom";
import { Layout } from "../components/Layout";
import { useAdminCreate } from "../components/AdminUI";
import { t } from "../i18n";
import "./admin.css";

const screens: Record<string, [string, string?]> = {
  members: ["Участники", "Добавить участника"],
  bookings: ["Записи", "Добавить запись"],
  attendance: ["Посещаемость"],
  settings: ["Настройки секции"],
  series: ["Серии", "Создать серию"],
  create: ["Создать занятие"],
  trainings: ["Создать занятие"],
  sections: ["Секции", "Создать секцию"],
  users: ["Пользователи"],
  administrators: ["Администраторы", "Назначить роль"],
};

export function AdminLayout() {
  const path = useLocation().pathname.split("/").filter(Boolean).pop() ?? "";
  const [label, add] = screens[path] ?? ["Админ"];
  const [params] = useSearchParams();
  const [, setOpen] = useAdminCreate();
  const header = {
    title: t(label),
    action: add && !(path === "administrators" && params.get("scope") === "system") && (
      <button className="btn primary" onClick={() => setOpen(true)}>
        + {t(add)}
      </button>
    ),
  };
  return (
    <Layout {...header}>
      <div className="admin-screen">
        <Outlet />
      </div>
    </Layout>
  );
}
