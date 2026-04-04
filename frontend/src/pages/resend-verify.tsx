import { useState } from "react";
import { Link } from "react-router-dom";

import { ApiError } from "../lib/api";
import { useAuth } from "../auth/use-auth";

export function ResendVerifyPage() {
  const { resendVerify } = useAuth();

  const [email, setEmail] = useState<string>("");
  const [msg, setMsg] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    setMsg("");
    setLoading(true);

    try {
      const message = await resendVerify(email);
      setMsg(
        `${message} backendログにverifyリンクが出るので、そのリンクを開いてください。`,
      );
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setMsg(err.message);
      } else {
        setMsg(
          "確認メールの再送に失敗しました。時間をおいて再度お試しください。",
        );
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto max-w-[960px]">
        <section className="mt-8 rounded-[32px] border border-[#eadfd5] bg-white px-6 py-7 shadow-[0_8px_24px_rgba(110,78,56,0.06)] md:px-8">
          <div className="mb-5">
            <span className="rounded-full bg-[#f4ebe3] px-3 py-1.5 text-[11px] font-black tracking-[0.28em] text-[#7b523a]">
              FORM
            </span>
          </div>

          <form onSubmit={onSubmit} className="grid gap-5">
            <div>
              <label
                htmlFor="email"
                className="mb-2 block text-sm font-black text-[#5d463a]"
              >
                メールアドレス
              </label>

              <input
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="user@example.com"
                type="email"
                autoComplete="email"
                className="w-full rounded-2xl border border-[#dccabf] bg-[#fffdfa] px-4 py-3 text-base font-semibold text-[#4e342e] outline-none transition placeholder:text-[#b19a8d] focus:border-[#8b5e3c] focus:ring-4 focus:ring-[#e7d6c7]"
              />
            </div>

            <div className="flex flex-wrap gap-3">
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center rounded-full bg-[#4e342e] px-5 py-3 text-sm font-bold text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {loading ? "送信中..." : "確認メールを再送する"}
              </button>

              <Link
                to="/login"
                className="inline-flex items-center rounded-full border border-[#d9c6b8] bg-white px-5 py-3 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
              >
                ログインへ
              </Link>
            </div>
          </form>

          {msg ? (
            <div className="mt-6 rounded-[24px] border border-[#eadfd5] bg-[#fcf7f2] px-5 py-4">
              <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                result
              </p>
              <p className="text-sm font-semibold leading-7 text-[#6d625b]">
                {msg}
              </p>
            </div>
          ) : null}
        </section>
      </div>
    </main>
  );
}
