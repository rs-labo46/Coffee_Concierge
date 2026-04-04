import { useState } from "react";
import { Link } from "react-router-dom";

import { ApiError } from "../lib/api";
import { useAuth } from "../auth/use-auth";
import { PasswordField } from "../components/password-field";

export function SignupPage() {
  const { signup } = useAuth();
  const [ok, setOk] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [msg, setMsg] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    setMsg("");
    setLoading(true);

    try {
      const message = await signup(email, password);
      setOk(true);
      setMsg(
        `${message} backendログにverifyリンクが出るので、そのリンクを開いてください。`,
      );
    } catch (err: unknown) {
      setOk(false);
      if (err instanceof ApiError) {
        setMsg(err.message);
      } else {
        setMsg("登録に失敗しました。");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-[calc(100vh-120px)] bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex w-full max-w-[1280px] justify-center">
        <section className="w-full max-w-[760px] rounded-[36px] border border-[#e6d9ce] bg-white px-6 py-8 shadow-[0_10px_28px_rgba(110,78,56,0.08)] md:px-8 md:py-10">
          <div className="mb-8">
            <p className="mb-3 text-sm font-black tracking-[0.28em] text-[#a1775b] uppercase">
              create account
            </p>

            <h2 className="mb-3 text-3xl font-black text-[#4e342e]">
              アカウント登録
            </h2>
          </div>

          {msg ? (
            <div
              className={[
                "mb-6 rounded-[24px] px-5 py-4 text-sm font-bold leading-7",
                ok
                  ? "border border-[#cfe4cc] bg-[#f3fff1] text-[#42613e]"
                  : "border border-[#e6c7bd] bg-[#fff3ef] text-[#8a4b3a]",
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

            <PasswordField
              label="パスワード"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="12文字以上のパスワード"
              autoComplete="new-password"
            />

            <button
              type="submit"
              disabled={loading}
              className="mt-2 inline-flex items-center justify-center rounded-2xl bg-[#4e342e] px-5 py-4 text-base font-black text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? "登録中..." : "アカウントを作成する"}
            </button>
          </form>

          <div className="mt-8 grid gap-3 rounded-[26px] border border-[#eadfd5] bg-[#fcf8f4] px-5 py-5">
            <p className="text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
              next action
            </p>

            <div className="flex flex-wrap gap-3">
              <Link
                to="/login"
                className="inline-flex rounded-full border border-[#d9c6b8] bg-white px-4 py-2.5 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
              >
                ログインへ
              </Link>

              <Link
                to="/"
                className="inline-flex rounded-full border border-[#d9c6b8] bg-white px-4 py-2.5 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
              >
                topへ戻る
              </Link>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
