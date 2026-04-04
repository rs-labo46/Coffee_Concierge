import { Link } from "react-router-dom";
import { useAuth } from "../auth/use-auth";

export function MePage() {
  const { user } = useAuth();

  if (!user) {
    return <div>no user</div>;
  }

  const roleTone =
    user.role === "admin"
      ? "bg-[#f1e3d6] text-[#7b523a]"
      : "bg-[#ece6ff] text-[#6a55aa]";

  return (
    <main className="min-h-[calc(100vh-120px)] bg-[#f6f1eb] px-4 py-8 md:px-8 md:py-10">
      <div className="mx-auto flex max-w-[1280px] flex-col gap-6">
        <section className="overflow-hidden rounded-[36px] border border-[#e6d9ce] bg-[#fffdfa] shadow-[0_10px_28px_rgba(110,78,56,0.08)]">
          <div className="border-b border-[#eadfd5] bg-gradient-to-r from-[#6f4e37] via-[#8b6448] to-[#ceb39e] px-8 py-10 text-white md:px-10">
            <p className="mb-3 text-sm font-black tracking-[0.32em] text-white/80 uppercase">
              my page
            </p>

            <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
              <div>
                <h1 className="mb-3 text-4xl font-black leading-tight md:text-5xl">
                  マイページ
                </h1>
              </div>

              <div className="flex flex-wrap gap-3">
                {user.role === "admin" ? (
                  <Link
                    to="/admin"
                    className="inline-flex rounded-full border border-white/30 bg-white/10 px-5 py-3 text-sm font-bold text-white transition hover:bg-white/20"
                  >
                    管理画面へ
                  </Link>
                ) : null}
              </div>
            </div>
          </div>

          <div className="grid gap-6 px-8 py-8 md:px-10 md:py-10 lg:grid-cols-[1.05fr_0.95fr]">
            <div className="grid gap-6">
              <section className="rounded-[30px] border border-[#eadfd5] bg-white px-6 py-6">
                <div className="mb-5 flex flex-wrap items-center gap-3">
                  <span
                    className={`rounded-full px-4 py-2 text-xs font-black tracking-[0.24em] uppercase ${roleTone}`}
                  >
                    {user.role}
                  </span>

                  <span className="rounded-full bg-[#f4ebe3] px-4 py-2 text-xs font-black tracking-[0.24em] text-[#7b523a] uppercase">
                    {user.email_verified ? "verified" : "not verified"}
                  </span>
                </div>

                <h2 className="mb-4 text-2xl font-black text-[#4e342e]">
                  アカウント概要
                </h2>

                <div className="grid gap-4 md:grid-cols-2">
                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      token version
                    </p>
                    <p className="text-xl font-black text-[#4e342e]">
                      {user.token_ver}
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      id
                    </p>
                    <p className="text-xl font-black text-[#4e342e]">
                      {user.id}
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5 md:col-span-2">
                    <p className="mb-2 text-xs font-black tracking-[0.24em] text-[#a1775b] uppercase">
                      email
                    </p>
                    <p className="break-all text-lg font-black text-[#4e342e]">
                      {user.email}
                    </p>
                  </div>
                </div>
              </section>
            </div>

            <div className="grid gap-6">
              <section className="rounded-[30px] border border-[#eadfd5] bg-white px-6 py-6">
                <p className="mb-3 text-sm font-black tracking-[0.24em] text-[#a1775b] uppercase">
                  アカウントのステータス
                </p>

                <h2 className="mb-4 text-2xl font-black text-[#4e342e]">
                  現在の状態
                </h2>

                <div className="grid gap-4">
                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-sm font-black text-[#5f4a40]">
                      ロール
                    </p>
                    <p className="text-base font-semibold leading-7 text-[#766b63]">
                      現在のロールは{" "}
                      <span className="font-black text-[#4e342e]">
                        {user.role}
                      </span>{" "}
                      です。
                    </p>
                  </div>

                  <div className="rounded-[24px] bg-[#fcf6f0] px-5 py-5">
                    <p className="mb-2 text-sm font-black text-[#5f4a40]">
                      メール確認
                    </p>
                    <p className="text-base font-semibold leading-7 text-[#766b63]">
                      状態は{" "}
                      <span className="font-black text-[#4e342e]">
                        {user.email_verified ? "確認済み" : "未確認"}
                      </span>{" "}
                      です。
                    </p>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
