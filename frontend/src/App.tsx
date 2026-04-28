import { type ReactNode } from "react";
import {
  BrowserRouter,
  NavLink,
  Navigate,
  Route,
  Routes,
  useNavigate,
} from "react-router-dom";
import { AuthProvider } from "./auth/auth";

import { AdminPage } from "./pages/admin";

import { ItemsPage } from "./pages/items";
import { LoginPage } from "./pages/login";
import { MePage } from "./pages/me";
import { ResendVerifyPage } from "./pages/resend-verify";

import { SignupPage } from "./pages/signup";
import TopPage from "./pages/top";
import { VerifyEmailPage } from "./pages/verify-email";
import { ForgotPasswordPage } from "./pages/forgot-password";
import { ItemDetailPage } from "./pages/item-detail";
import { ResetPasswordPage } from "./pages/reset-password";
import { ConciergePage } from "./pages/concierge";
import { useAuth } from "./auth/use-auth";

type GuardProps = {
  children: ReactNode;
};

function GuardLoading() {
  return (
    <div className="mx-auto max-w-[1280px] px-4 py-10 md:px-8">
      <div className="rounded-[28px] border border-[#e6d9ce] bg-white px-6 py-10 text-center text-base font-bold text-[#6f6259] shadow-[0_8px_24px_rgba(110,78,56,0.06)]">
        読み込み中です...
      </div>
    </div>
  );
}

function RequireAuth({ children }: GuardProps) {
  const { user, loading } = useAuth();

  if (loading) {
    return <GuardLoading />;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function RequireAdmin({ children }: GuardProps) {
  const { user, loading } = useAuth();

  if (loading) {
    return <GuardLoading />;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (user.role !== "admin") {
    return <Navigate to="/me" replace />;
  }

  return <>{children}</>;
}

function HeaderNavItem({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        [
          "inline-flex min-h-12 items-center justify-center rounded-full border px-4 py-2 text-sm font-bold transition",
          isActive
            ? "border-[#7b523a] bg-[#7b523a] text-white shadow-[0_8px_20px_rgba(123,82,58,0.22)]"
            : "border-[#d9c6b8] bg-white text-[#7b523a] hover:-translate-y-0.5 hover:bg-[#f8efe7]",
        ].join(" ")
      }
    >
      {label}
    </NavLink>
  );
}

function HeaderAuthActions() {
  const { user, logout } = useAuth();
  const nav = useNavigate();

  async function onLogout() {
    try {
      await logout();
    } finally {
      nav("/login", { replace: true });
    }
  }

  if (!user) {
    return (
      <div className="flex w-full flex-wrap items-center justify-center gap-2 lg:w-auto lg:justify-end">
        <NavLink
          to="/login"
          className={({ isActive }) =>
            [
              "inline-flex min-h-12 items-center justify-center rounded-full border px-4 py-2 text-sm font-bold transition",
              isActive
                ? "border-[#7b523a] bg-[#7b523a] text-white"
                : "border-[#d9c6b8] bg-white text-[#7b523a] hover:bg-[#f8efe7]",
            ].join(" ")
          }
        >
          ログイン
        </NavLink>

        <NavLink
          to="/signup"
          className="inline-flex min-h-12 items-center justify-center rounded-full bg-[#4e342e] px-4 py-2 text-sm font-bold text-white transition hover:opacity-90"
        >
          新規登録
        </NavLink>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-3 lg:w-auto lg:flex-row lg:items-center lg:justify-end">
      <div className="flex w-full justify-center lg:w-auto lg:justify-end">
        <div className="flex min-h-12 w-full max-w-[460px] min-w-0 items-center justify-center gap-2 rounded-full border border-[#e2d3c7] bg-[#fbf7f2] px-4 py-2 lg:w-auto lg:justify-start">
          <span className="rounded-full bg-[#f1e3d6] px-2.5 py-1 text-[11px] font-black tracking-[0.2em] text-[#7b523a] uppercase">
            {user.role}
          </span>

          <span className="min-w-0 truncate text-sm font-bold text-[#5f4a40]">
            {user.email}
          </span>
        </div>
      </div>

      <div className="flex w-full justify-center lg:w-auto lg:justify-end">
        <button
          type="button"
          onClick={() => void onLogout()}
          className="inline-flex min-h-12 w-full max-w-[300px] items-center justify-center rounded-full bg-[#4e342e] px-5 py-2 text-sm font-bold text-white transition hover:opacity-90 lg:w-auto"
        >
          ログアウト
        </button>
      </div>
    </div>
  );
}

function AppHeader() {
  const { user } = useAuth();

  return (
    <header className="sticky top-0 z-50 border-b border-[#e6d9ce]/80 bg-[#fffaf5]/90 backdrop-blur">
      <div className="h-1 w-full bg-gradient-to-r from-[#6f4e37] via-[#b08968] to-[#e7d6c7]" />

      <div className="mx-auto max-w-[1440px] px-4 py-4 md:px-8">
        <div className="rounded-[28px] border border-[#e6d9ce] bg-white/95 px-4 py-4 shadow-[0_10px_28px_rgba(110,78,56,0.08)] md:px-6">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex justify-center lg:justify-start">
              <NavLink
                to="/"
                className="flex w-full min-w-0 items-center justify-center gap-3 rounded-[22px] border border-[#eadfd5] bg-[#fcf7f2] px-3 py-3 text-center transition hover:bg-[#f7efe8] sm:w-fit sm:px-5 lg:justify-start lg:text-left"
              >
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-[#6f4e37] text-lg font-black tracking-[0.12em] text-white">
                  CS
                </div>

                <div className="min-w-0 text-center lg:text-left">
                  <h1 className="truncate text-lg font-black text-[#4e342e] md:text-xl">
                    coffee spa
                  </h1>
                  <p className="hidden text-sm font-semibold text-[#8a7b71] md:block">
                    豆・レシピ・セール・店舗情報を一つに。
                  </p>
                </div>
              </NavLink>
            </div>

            <div className="flex w-full flex-col gap-3 lg:w-auto lg:items-end">
              <nav className="flex w-full flex-wrap items-center justify-center gap-2 lg:w-auto lg:justify-end">
                <HeaderNavItem to="/concierge" label="コンシェルジュ" />

                {user ? <HeaderNavItem to="/me" label="マイページ" /> : null}
                {user?.role === "admin" ? (
                  <HeaderNavItem to="/admin" label="管理" />
                ) : null}
              </nav>
              ;
              <HeaderAuthActions />
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

function NotFoundPage() {
  return (
    <div className="mx-auto max-w-[1280px] px-4 py-10 md:px-8">
      <div className="rounded-[32px] border border-[#eadfd4] bg-white px-8 py-16 text-center shadow-[0_8px_24px_rgba(110,78,56,0.06)]">
        <p className="mb-3 text-sm font-black tracking-[0.28em] text-[#a1775b] uppercase">
          404
        </p>
        <h2 className="mb-4 text-3xl font-black text-[#4e342e]">
          page not found
        </h2>
        <p className="text-base font-semibold text-[#7a6f68]">
          指定されたページは見つかりませんでした。
        </p>
      </div>
    </div>
  );
}

function AppRoutes() {
  return (
    <div className="min-h-screen bg-[#f6f1eb]">
      <AppHeader />

      <main className="pb-10">
        <Routes>
          <Route path="/" element={<TopPage />} />
          <Route path="/items" element={<ItemsPage />} />
          <Route path="/items/:id" element={<ItemDetailPage />} />
          <Route path="/concierge" element={<ConciergePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/verify-email" element={<VerifyEmailPage />} />
          <Route path="/resend-verify" element={<ResendVerifyPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
          <Route path="/reset-password" element={<ResetPasswordPage />} />

          <Route
            path="/me"
            element={
              <RequireAuth>
                <MePage />
              </RequireAuth>
            }
          />

          <Route
            path="/admin"
            element={
              <RequireAdmin>
                <AdminPage />
              </RequireAdmin>
            }
          />

          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  );
}
