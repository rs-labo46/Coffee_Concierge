import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";

import { ApiError } from "../lib/api";
import { useAuth } from "../auth/use-auth";

type VerifyState = "loading" | "success" | "error";

function VerifyTone(state: VerifyState) {
  switch (state) {
    case "success":
      return {
        badge: "SUCCESS",
        badgeClass: "bg-[#e8f5ea] text-[#2f6b3d]",
        panelClass: "border-[#d6eadb] bg-[#f8fcf9]",
        title: "メール確認が完了しました",
        titleClass: "text-[#2f5d39]",
      };
    case "error":
      return {
        badge: "ERROR",
        badgeClass: "bg-[#fbe8e6] text-[#9a463d]",
        panelClass: "border-[#efd8d3] bg-[#fffaf8]",
        title: "メール確認に失敗しました",
        titleClass: "text-[#8a433a]",
      };
    default:
      return {
        badge: "VERIFY",
        badgeClass: "bg-[#f4ebe3] text-[#7b523a]",
        panelClass: "border-[#eadfd5] bg-[#fffdfa]",
        title: "確認処理を実行しています",
        titleClass: "text-[#4e342e]",
      };
  }
}

export function VerifyEmailPage() {
  const { verifyEmail } = useAuth();
  const [params] = useSearchParams();

  const [msg, setMsg] = useState<string>("メール確認を進めています...");
  const [state, setState] = useState<VerifyState>("loading");

  // 同じtokenで effect が二重実行されても、
  // 1回しかAPIを叩かないようにする。
  const startedRef = useRef<boolean>(false);

  const token = params.get("token") || "";
  const tone = useMemo(() => VerifyTone(state), [state]);

  useEffect(() => {
    if (startedRef.current) {
      return;
    }
    startedRef.current = true;

    async function run() {
      if (!token) {
        setState("error");
        setMsg(
          "確認用トークンが見つかりません。メールのリンクを開き直してください。",
        );
        return;
      }

      try {
        await verifyEmail(token);
        setState("success");
        setMsg(
          "メール確認が完了しました。ログインして利用を開始してください。",
        );
      } catch (err: unknown) {
        setState("error");

        if (err instanceof ApiError) {
          setMsg(err.message);
          return;
        }

        setMsg("メール確認に失敗しました。時間をおいて再度お試しください。");
      }
    }

    void run();
  }, [token, verifyEmail]);

  return (
    <main className="min-h-[calc(100vh-120px)] bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex w-full max-w-[1280px] justify-center">
        <section
          className={`w-full max-w-[760px] rounded-[32px] border px-6 py-7 shadow-[0_8px_24px_rgba(110,78,56,0.06)] md:px-8 ${tone.panelClass}`}
        >
          <div className="mb-4 flex flex-wrap items-center gap-3">
            <span
              className={`rounded-full px-3 py-1.5 text-[11px] font-black tracking-[0.28em] ${tone.badgeClass}`}
            >
              {tone.badge}
            </span>

            {state === "loading" ? (
              <span className="text-sm font-bold text-[#8b7a6c]">
                処理を実行中です
              </span>
            ) : null}
          </div>

          <h2 className={`mb-3 text-2xl font-black ${tone.titleClass}`}>
            {tone.title}
          </h2>

          <p className="max-w-3xl text-base font-semibold leading-8 text-[#6d625b]">
            {msg}
          </p>

          <div className="mt-6 flex flex-wrap gap-3">
            {state === "success" ? (
              <>
                <Link
                  to="/login"
                  className="inline-flex items-center rounded-full bg-[#4e342e] px-5 py-3 text-sm font-bold text-white transition hover:opacity-90"
                >
                  ログインへ
                </Link>

                <Link
                  to="/"
                  className="inline-flex items-center rounded-full border border-[#d9c6b8] bg-white px-5 py-3 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
                >
                  トップへ
                </Link>
              </>
            ) : state === "error" ? (
              <>
                <Link
                  to="/resend-verify"
                  className="inline-flex items-center rounded-full bg-[#4e342e] px-5 py-3 text-sm font-bold text-white transition hover:opacity-90"
                >
                  確認メールを再送する
                </Link>

                <Link
                  to="/login"
                  className="inline-flex items-center rounded-full border border-[#d9c6b8] bg-white px-5 py-3 text-sm font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
                >
                  ログインへ戻る
                </Link>
              </>
            ) : (
              <div className="inline-flex items-center rounded-full border border-[#eadfd5] bg-white px-5 py-3 text-sm font-bold text-[#8b7a6c]">
                少し待ってください...
              </div>
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
