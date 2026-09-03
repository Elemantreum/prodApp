import type { ReactNode } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider, useAuth } from "./auth/AuthContext";
import { LoginGate } from "./pages/LoginGate";
import { RegistrationGate } from "./pages/RegistrationGate";
import { TrainingsPage } from "./pages/TrainingsPage";
import { MyBookingsPage } from "./pages/MyBookingsPage";
import { AdminLayout } from "./pages/AdminLayout";
import { AdminUsersPage } from "./pages/AdminUsersPage";
import { AdminCreateTrainingPage } from "./pages/AdminCreateTrainingPage";
import { AccountPage } from "./pages/AccountPage";

function AdminRoute({ children }: { children: ReactNode }) {
  const { claims } = useAuth();
  if (!claims?.is_admin) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <LoginGate>
          <RegistrationGate>
            <Routes>
              <Route path="/" element={<TrainingsPage />} />
              <Route path="/bookings" element={<MyBookingsPage />} />
              <Route path="/account" element={<AccountPage />} />
              <Route
                path="/admin"
                element={
                  <AdminRoute>
                    <AdminLayout />
                  </AdminRoute>
                }
              >
                <Route index element={<Navigate to="users" replace />} />
                <Route path="users" element={<AdminUsersPage />} />
                <Route path="trainings" element={<AdminCreateTrainingPage />} />
              </Route>
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </RegistrationGate>
        </LoginGate>
      </BrowserRouter>
    </AuthProvider>
  );
}
