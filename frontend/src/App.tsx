import { AdminMembersPage, AdminAdministratorsPage } from "./pages/AdminMembersPage";
import { PersonalAttendancePage } from "./pages/PersonalAttendancePage";
import { t as translate } from "./i18n";
import { useEffect, useState, type ReactNode } from "react";
import { BrowserRouter, Link, Navigate, Outlet, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import { SectionProvider, useSection, useSectionDirectory } from "./auth/Sections";
import { AdminSectionsPage, SectionSettings } from "./pages/SectionSettings";
import { AuthProvider } from "./auth/AuthContext";
import { BookingHistory } from "./components/BookingHistory";
import { Layout } from "./components/Layout";
import { LoginGate } from "./pages/LoginGate";
import { RegistrationGate } from "./pages/RegistrationGate";
import { TrainingsPage } from "./pages/TrainingsPage";
import { MyBookingsPage } from "./pages/MyBookingsPage";
import { SinnersPage } from "./pages/SinnersPage";
import { AdminLayout } from "./pages/AdminLayout";
import { AdminUsersPage } from "./pages/AdminUsersPage";
import { AdminCreateTrainingPage, AdminSeriesPage } from "./pages/AdminCreateTrainingPage";
import { AccountPage } from "./pages/AccountPage";
import { InvitePage } from "./pages/InvitePage";
import { getTelegramInviteCode } from "./platform/telegram";

function TelegramInviteLaunch() {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const [code, setCode] = useState(() => (pathname === "/" ? getTelegramInviteCode() : undefined));

  useEffect(() => {
    if (!code) return;
    navigate({ pathname: "/invite", hash: `#${new URLSearchParams({ invite: code })}` }, { replace: true });
    setCode(undefined);
  }, [code, navigate]);
  return null;
}

function SectionRoutes() {
  const { current } = useSectionDirectory();
  return current ? <Outlet key={current.id} /> : <Navigate to="/invite" replace />;
}

function AdminRoute({ children, system = false }: { children: ReactNode; system?: boolean }) {
  const { current, directory } = useSection();
  if (!current.admin || (system && !directory.systemAdmin)) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <TelegramInviteLaunch />
        <LoginGate>
          <RegistrationGate>
            <SectionProvider>
              <Routes>
                <Route path="/invite" element={<InvitePage />} />
                <Route path="/account" element={<AccountPage />} />
                <Route element={<SectionRoutes />}>
                  <Route path="/" element={<TrainingsPage />} />
                  <Route path="/bookings" element={<MyBookingsPage />} />
                  <Route path="/attendance" element={<PersonalAttendancePage />} />
                  <Route path="/sinners" element={<SinnersPage />} />
                  <Route
                    path="/admin"
                    element={
                      <AdminRoute>
                        <AdminLayout />
                      </AdminRoute>
                    }
                  >
                    <Route index element={<Navigate to="bookings" replace />} />
                    <Route path="members" element={<AdminMembersPage />} />
                    <Route path="series" element={<AdminSeriesPage />} />
                    <Route path="create" element={<AdminCreateTrainingPage />} />
                    <Route
                      path="sections"
                      element={
                        <AdminRoute system>
                          <AdminSectionsPage />
                        </AdminRoute>
                      }
                    />
                    <Route
                      path="administrators"
                      element={
                        <AdminRoute system>
                          <AdminAdministratorsPage />
                        </AdminRoute>
                      }
                    />
                    <Route path="bookings" element={<BookingHistory />} />
                    <Route path="attendance" element={<BookingHistory attendance />} />
                    <Route path="settings" element={<SectionSettings />} />
                    <Route
                      path="users"
                      element={
                        <AdminRoute system>
                          <AdminUsersPage />
                        </AdminRoute>
                      }
                    />
                    <Route path="trainings" element={<AdminCreateTrainingPage />} />
                  </Route>
                  <Route
                    path="*"
                    element={
                      <Layout title={translate("Страница не найдена")}>
                        <Link to="/">{translate("Вернуться к расписанию")}</Link>
                      </Layout>
                    }
                  />
                </Route>
              </Routes>
            </SectionProvider>
          </RegistrationGate>
        </LoginGate>
      </BrowserRouter>
    </AuthProvider>
  );
}
