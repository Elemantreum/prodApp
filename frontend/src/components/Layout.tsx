import { useState } from "react";
import type { ReactNode } from "react";
import { Link, useLocation, useSearchParams } from "react-router-dom";
import * as Dialog from "@radix-ui/react-dialog";
import { useSection } from "../auth/Sections";
import { useProfile } from "../auth/Profile";
import { useAuth } from "../auth/AuthContext";
import { applyTheme } from "../platform/theme";
import { t } from "../i18n";
import { isSystemPage, Navigation, SystemNavigation } from "./TabBar";
import { LanguageSelect } from "./LanguageSelect";

function Brand() {
  return (
    <Link className="brand" to="/">
      <span className="logo" aria-hidden="true">
        p
      </span>
      ProdApp
    </Link>
  );
}

function SectionPicker() {
  const { current, directory, choose } = useSection();
  const { pathname } = useLocation();
  const [, setParams] = useSearchParams();
  return (
    <label className="section-picker">
      {current.logo ? (
        <img src={current.logo} alt="" width="22" height="22" />
      ) : (
        <span aria-hidden="true">◈</span>
      )}
      <select
        className="scope"
        value={current.id}
        aria-label={t("Секция")}
        onChange={(event) => {
          if (pathname === "/")
            setParams(
              (previous) => {
                const next = new URLSearchParams(previous);
                next.delete("session");
                next.delete("new");
                return next;
              },
              { replace: true },
            );
          choose(Number(event.target.value));
        }}
      >
        {directory.sections.map((s) => (
          <option key={s.id} value={s.id}>
            {s.name}
          </option>
        ))}
      </select>
    </label>
  );
}

interface LayoutProps {
  title: string;
  children: ReactNode;
  action?: ReactNode;
  publicPage?: boolean;
  standalone?: boolean;
}

export function Layout(props: LayoutProps) {
  return props.publicPage || props.standalone ? <StandaloneLayout {...props} /> : <MemberLayout {...props} />;
}

function StandaloneLayout({ title, children, publicPage }: LayoutProps) {
  const { platform } = useAuth();
  const { pathname } = useLocation();
  const role = publicPage ? t("Гость") : null;
  const showProfile = !publicPage && pathname !== "/account";
  return (
    <div className={platform === "browser" ? "app-frame" : "app-frame compact"}>
      <div className="shell">
        <aside className="side">
          <Brand />
          <nav className="nav">
            {publicPage ? (
              <span className="public-tab" aria-current="page">
                {title}
              </span>
            ) : (
              <>
                <Link to="/invite">{t("Ввести код приглашения")}</Link>
                <Link to="/account">{t("Профиль")}</Link>
              </>
            )}
          </nav>
          {role && <div className="sidefoot">{role}</div>}
        </aside>
        <main>
          <div className="mobilebrand">
            <Brand />
            {showProfile && (
              <Link className="btn" to="/account">
                {t("Профиль")}
              </Link>
            )}
          </div>
          <header className="top">
            {role && <span>{role}</span>}
            <LanguageSelect />
          </header>
          <div className="wrap">
            <div className="head">
              <div>
                {role && <div className="eyebrow">{role}</div>}
                <h1>{title}</h1>
              </div>
            </div>
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}

function MemberLayout({ title, children, action }: LayoutProps) {
  const { current, directory } = useSection();
  const { profile } = useProfile();
  const { platform } = useAuth();
  const { pathname } = useLocation();
  const [open, setOpen] = useState(false);
  const [theme, setTheme] = useState(localStorage.getItem("prodapp.theme") ?? "auto");
  const system = directory.systemAdmin && isSystemPage(pathname);
  const role = t(
    system ? "Системный администратор" : current.admin ? "Администратор секции" : "Участник секции",
  );
  const name = [profile.firstName, profile.lastName].filter(Boolean).join(" ");
  const initials = [profile.firstName, profile.lastName]
    .map((part) => part?.[0])
    .join("")
    .toUpperCase();
  return (
    <div className={platform === "browser" ? "app-frame" : "app-frame compact"}>
      <div className="shell">
        <aside className="side">
          <Brand />
          <SectionPicker />
          <Navigation />
          {directory.systemAdmin ? (
            <SystemNavigation onMore={() => setOpen(true)} />
          ) : (
            <div className="sidefoot">
              <span>{role}</span>
              <button aria-label={t("Меню")} onClick={() => setOpen(true)}>
                ···
              </button>
            </div>
          )}
        </aside>
        <main>
          <div className="mobilebrand">
            <Brand />
            <button className="btn" onClick={() => setOpen(true)}>
              {t("Меню")}
            </button>
          </div>
          <header className="top">
            <span>{system ? t("Управление системой") : current.name}</span>
            <Link to="/account" className="top-profile">
              {name}
              <span className="avatar">{initials}</span>
            </Link>
          </header>
          <div className="wrap">
            {!system && (
              <div className="mobile-scope">
                <SectionPicker />
              </div>
            )}
            <div className="head">
              <div>
                <div className="eyebrow">{pathname === "/" ? current.name : role}</div>
                <h1>{title}</h1>
              </div>
              {action}
            </div>
            {children}
          </div>
        </main>
      </div>
      <Navigation compact onMore={() => setOpen(true)} />
      <Dialog.Root open={open} onOpenChange={setOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="shell-shade" />
          <Dialog.Content className="shell-dialog" aria-describedby={undefined}>
            <div className="bar">
              <Dialog.Title>{t("Меню")}</Dialog.Title>
              <Dialog.Close className="btn" aria-label={t("Закрыть")}>
                ×
              </Dialog.Close>
            </div>
            <Navigation onNavigate={() => setOpen(false)} />
            <Link className="shell-extra" to="/invite" onClick={() => setOpen(false)}>
              {t("Ввести код приглашения")}
            </Link>
            <Link className="shell-extra" to="/sinners" onClick={() => setOpen(false)}>
              {t("Список грешников")}
            </Link>
            <SystemNavigation onNavigate={() => setOpen(false)} />
            <div className="shell-preferences">
              <label>
                {t("Тема")}{" "}
                <select
                  value={theme}
                  onChange={(e) => {
                    setTheme(e.target.value);
                    localStorage.setItem("prodapp.theme", e.target.value);
                    applyTheme();
                  }}
                >
                  <option value="auto">{t("Системная")}</option>
                  <option value="dark">{t("Тёмная")}</option>
                  <option value="light">{t("Светлая")}</option>
                </select>
              </label>
              <LanguageSelect />
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
