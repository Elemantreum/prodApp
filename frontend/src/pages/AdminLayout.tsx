import { NavLink, Outlet } from "react-router-dom";
import { Layout } from "../components/Layout";

export function AdminLayout() {
  return (
    <Layout title="Админ">
      <div className="segmented">
        <NavLink to="/admin/users" className={({ isActive }) => `segmented__item${isActive ? " segmented__item--active" : ""}`}>
          Пользователи
        </NavLink>
        <NavLink
          to="/admin/trainings"
          className={({ isActive }) => `segmented__item${isActive ? " segmented__item--active" : ""}`}
        >
          Тренировки
        </NavLink>
      </div>
      <Outlet />
    </Layout>
  );
}
