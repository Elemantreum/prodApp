import { matchPath, NavLink } from "react-router-dom";
import { useSection } from "../auth/Sections";
import { t } from "../i18n";

const schedule = ["/", "Расписание", "▦"] as const;
const bookings = ["/bookings", "Мои записи", "▤"] as const;
const profile = ["/account", "Профиль", "◉"] as const;
type Destination = readonly [path: string, label: string, icon: string];

const systemPages: Destination[] = [
  ["/admin/sections", "Секции", "◈"],
  ["/admin/users", "Пользователи", "◎"],
  ["/admin/administrators", "Администраторы", "◇"],
];

export function isSystemPage(pathname: string) {
  return systemPages.some(([path]) => matchPath(path, pathname));
}

function NavigationLinks({ pages, onNavigate }: { pages: Destination[]; onNavigate?: () => void }) {
  return pages.map(([path, label, icon]) => (
    <NavLink
      key={path}
      to={path}
      end
      onClick={onNavigate}
      className={({ isActive }) => (isActive ? "current" : "")}
    >
      <span className="num" aria-hidden="true">
        {icon}
      </span>
      {t(label)}
    </NavLink>
  ));
}

export function SystemNavigation({ onMore, onNavigate }: { onMore?: () => void; onNavigate?: () => void }) {
  const { directory } = useSection();
  if (!directory.systemAdmin) return null;
  const title = t("Управление системой");

  return (
    <section className="system-menu">
      <div className="system-menu__header">
        <h2>{title}</h2>
        {onMore && (
          <button aria-label={t("Меню")} onClick={onMore}>
            ···
          </button>
        )}
      </div>
      <nav className="nav" aria-label={title}>
        <NavigationLinks pages={systemPages} onNavigate={onNavigate} />
      </nav>
    </section>
  );
}

export function Navigation({
  compact = false,
  onMore,
  onNavigate,
}: {
  compact?: boolean;
  onMore?: () => void;
  onNavigate?: () => void;
}) {
  const { current } = useSection();
  const pages: Destination[] = [schedule, bookings];
  if (!compact) {
    if (current.admin) {
      pages.push(
        ["/admin/members", "Участники", "◎"],
        ["/admin/bookings", "Записи", "▤"],
        ["/admin/series", "Серии", "↻"],
        ["/admin/attendance", "Посещаемость", "◷"],
        ["/admin/settings", "Настройки секции", "⚙"],
      );
    } else pages.push(["/attendance", "Посещаемость", "◷"]);
  }
  pages.push(profile);
  return (
    <nav className={compact ? "mobileNav" : "nav"} aria-label={t("Навигация")}>
      <NavigationLinks pages={pages} onNavigate={onNavigate} />
      {compact && (
        <button onClick={onMore}>
          <span className="num" aria-hidden="true">
            ···
          </span>
          {t("Ещё")}
        </button>
      )}
    </nav>
  );
}
