import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  ApiError,
  api,
  clearToken,
  getCookie,
  getToken,
  setToken,
} from "../lib/api";
import { AuthContext, type AuthCtx, type User } from "./context";

type SignupResponse = {
  user: User;
};

type LoginResponse = {
  access_token: string;
  user: User;
};

type RefreshResponse = {
  access_token: string;
  user: User;
};

type MeResponse = {
  user: User;
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const initStartedRefresh = useRef<boolean>(false);

  async function signup(email: string, password: string): Promise<string> {
    const res = await api<SignupResponse>("/auth/signup", {
      method: "POST",
      body: { email, password },
    });

    if (!res) {
      return "登録しました。メール確認をしてください。";
    }

    return `${res.user.email} を登録しました。メール確認をしてください。`;
  }

  async function verifyEmail(token: string): Promise<void> {
    await api<void>("/auth/verify-email", {
      method: "POST",
      body: { token },
    });
  }

  async function resendVerify(email: string): Promise<string> {
    await api<void>("/auth/resend-verify", {
      method: "POST",
      body: { email },
    });

    return `${email} 宛に確認メールを再送しました。`;
  }

  async function forgotPassword(email: string): Promise<string> {
    await api<void>("/auth/password/forgot", {
      method: "POST",
      body: { email },
    });

    return "入力したメールアドレス宛に、再設定案内を送信しました。";
  }

  async function resetPassword(
    token: string,
    newPassword: string,
  ): Promise<void> {
    await api<void>("/auth/password/reset", {
      method: "POST",
      body: {
        token,
        password: newPassword,
      },
    });
  }

  async function loadMe(): Promise<void> {
    const res = await api<MeResponse>("/me", {
      method: "GET",
      auth: true,
    });

    if (!res) {
      throw new Error("me response is empty");
    }

    setUser(res.user);
  }

  async function login(email: string, password: string): Promise<void> {
    const res = await api<LoginResponse>("/auth/login", {
      method: "POST",
      body: { email, password },
    });

    if (!res) {
      throw new Error("login response is empty");
    }

    setToken(res.access_token);
    setUser(res.user);
  }

  async function refresh(): Promise<boolean> {
    try {
      const res = await api<RefreshResponse>("/auth/refresh", {
        method: "POST",
        csrf: true,
      });

      if (!res) {
        clearToken();
        setUser(null);
        return false;
      }

      setToken(res.access_token);
      setUser(res.user);
      return true;
    } catch (err: unknown) {
      clearToken();
      setUser(null);

      if (err instanceof ApiError) {
        if (err.status === 401 || err.status === 403) {
          return false;
        }
      }

      return false;
    }
  }

  async function logout() {
    try {
      await api("/auth/logout", {
        method: "POST",
        auth: true,
        csrf: true,
      });
    } catch (err: unknown) {
      if (!(err instanceof ApiError && err.status === 401)) {
        throw err;
      }
    } finally {
      clearToken();
      setUser(null);
    }
  }

  useEffect(() => {
    if (initStartedRefresh.current) {
      return;
    }

    initStartedRefresh.current = true;

    async function init() {
      try {
        const token = getToken();
        if (token) {
          try {
            await loadMe();
            return;
          } catch {
            clearToken();
            setUser(null);
          }
        }

        const csrf = getCookie("csrf_token");
        if (!csrf) {
          setUser(null);
          return;
        }

        await refresh();
      } finally {
        setLoading(false);
      }
    }

    void init();
  }, []);

  const value = useMemo<AuthCtx>(() => {
    return {
      user,
      loading,
      signup,
      verifyEmail,
      resendVerify,
      forgotPassword,
      resetPassword,
      login,
      refresh,
      logout,
      loadMe,
    };
  }, [user, loading]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
