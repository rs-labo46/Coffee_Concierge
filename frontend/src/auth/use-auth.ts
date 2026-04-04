import { useContext } from "react";
import { AuthContext, type AuthCtx } from "./context";

export function useAuth(): AuthCtx {
  const ctx = useContext(AuthContext);

  if (!ctx) {
    throw new Error("auth context not found");
  }

  return ctx;
}
