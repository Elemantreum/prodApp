import { NavLink } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

const ICONS = {
  trainings: (
    <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
      <path d="M4 7h16M4 12h16M4 17h10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  ),
  bookings: (
    <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
      <rect x="4" y="5" width="16" height="15" rx="2" stroke="currentColor" strokeWidth="2" />
      <path d="M8 3v4M16 3v4M4 10h16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  ),
  admin: (
    <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
      <path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6l7-3z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" />
    </svg>
  ),
  sinners: (
    <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
      <circle cx="12" cy="9" r="5" stroke="currentColor" strokeWidth="2" />
      <path d="M9.5 8h.01M14.5 8h.01M9 12c1 1 4 1 5 0" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <path d="M8 13.5L5 21M16 13.5l3 7.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  ),
  account: (
    <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
      <circle cx="12" cy="8" r="3.5" stroke="currentColor" strokeWidth="2" />
      <path d="M4.5 20c1.5-4 5-5.5 7.5-5.5s6 1.5 7.5 5.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  ),
};

export function TabBar() {
  const { claims } = useAuth();

  return (
    <nav className="tab-bar">
      <NavLink to="/" end className={({ isActive }) => `tab-bar__item${isActive ? " tab-bar__item--active" : ""}`}>
        {ICONS.trainings}
        <span>Тренировки</span>
      </NavLink>
      <NavLink to="/bookings" className={({ isActive }) => `tab-bar__item${isActive ? " tab-bar__item--active" : ""}`}>
        {ICONS.bookings}
        <span>Мои записи</span>
      </NavLink>
      <NavLink to="/sinners" className={({ isActive }) => `tab-bar__item${isActive ? " tab-bar__item--active" : ""}`}>
        {ICONS.sinners}
        <span>Грешники</span>
      </NavLink>
      {claims?.is_admin && (
        <NavLink to="/admin" className={({ isActive }) => `tab-bar__item${isActive ? " tab-bar__item--active" : ""}`}>
          {ICONS.admin}
          <span>Админ</span>
        </NavLink>
      )}
      <NavLink to="/account" className={({ isActive }) => `tab-bar__item${isActive ? " tab-bar__item--active" : ""}`}>
        {ICONS.account}
        <span>Аккаунт</span>
      </NavLink>
    </nav>
  );
}
