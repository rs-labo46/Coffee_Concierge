import { useState } from "react";
import { Link } from "react-router-dom";
import { ApiError } from "../lib/api";
import { useAuth } from "../auth/use-auth";

export function ForgotPasswordPage() {
  const { forgotPassword } = useAuth();
  const [email, setEmail] = useState("");
  const [msg, setMsg] = useState("");
  const [done, setDone] = useState(false);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setMsg("");
    setLoading(true);

    try {
      const message = await forgotPassword(email);
      setMsg(message);
      setDone(true);
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setMsg(err.message);
      } else {
        setMsg("再設定メールの送信に失敗しました。");
      }
      setDone(false);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto max-w-[960px]">
        <section className="rounded-[36px] border border-[#e6d9ce] bg-white px-6 py-8 shadow-[0_10px_28px_rgba(110,78,56,0.08)] md:px-8 md:py-10">
          <p className="mb-3 text-sm font-black tracking-[0.28em] text-[#a1775b] uppercase">
            password reset
          </p>

          <h1 className="mb-3 text-3xl font-black text-[#4e342e]">
            パスワード再設定メールを送る
          </h1>

          <p className="mb-8 text-base font-semibold leading-8 text-[#766b63]">
            登録済みメールアドレスを入力してください。該当アカウントがあれば再設定リンクを送信します。
          </p>

          {msg ? (
            <div
              className={[
                "mb-6 rounded-[24px] border px-5 py-4 text-sm font-bold leading-7",
                done
                  ? "border-[#d6eadb] bg-[#f8fcf9] text-[#2f6b3d]"
                  : "border-[#e6c7bd] bg-[#fff3ef] text-[#8a4b3a]",
              ].join(" ")}
            >
              {msg}
            </div>
          ) : null}

          <form onSubmit={onSubmit} className="grid gap-5">
            <label className="grid gap-2">
              <span className="text-sm font-black tracking-[0.08em] text-[#5f4a40]">
                メールアドレス
              </span>
              <input
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                type="email"
                autoComplete="email"
                className="w-full rounded-2xl border border-[#dccabc] bg-[#fffdfa] px-4 py-4 text-base font-semibold text-[#4e342e] outline-none transition placeholder:text-[#b09d90] focus:border-[#9c7257] focus:ring-4 focus:ring-[#ead8ca]"
              />
            </label>

            <button
              type="submit"
              disabled={loading}
              className="inline-flex items-center justify-center rounded-2xl bg-[#4e342e] px-5 py-4 text-base font-black text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? "送信中..." : "再設定メールを送信"}
            </button>
          </form>

          <div className="mt-8 flex flex-wrap gap-3">
            <Link
              to="/login"
              className="inline-flex rounded-full border border-[#d9c6b8] bg-white px-4 py-2.5 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
            >
              ログインへ戻る
            </Link>
          </div>
        </section>
      </div>
    </main>
  );
}
