import { createContext } from "react";

export type Role = "user" | "admin";

export type User = {
  id: number;
  email: string;
  role: Role;
  token_ver: number;
  email_verified: boolean;
};

export type AuthCtx = {
  user: User | null;
  loading: boolean;
  signup: (email: string, password: string) => Promise<string>;
  verifyEmail: (token: string) => Promise<void>;
  resendVerify: (email: string) => Promise<string>;
  forgotPassword: (email: string) => Promise<string>;
  resetPassword: (token: string, newPassword: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  refresh: () => Promise<boolean>;
  logout: () => Promise<void>;
  loadMe: () => Promise<void>;
};

export const AuthContext = createContext<AuthCtx | null>(null);
