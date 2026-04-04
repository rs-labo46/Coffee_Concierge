import { useState } from "react";

type Props = {
  label: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder: string;
  autoComplete?: string;
};

export function PasswordField({
  label,
  value,
  onChange,
  placeholder,
  autoComplete,
}: Props) {
  const [show, setShow] = useState(false);

  return (
    <label className="grid gap-2">
      <span className="text-sm font-black tracking-[0.08em] text-[#5f4a40]">
        {label}
      </span>

      <div className="relative">
        <input
          value={value}
          onChange={onChange}
          placeholder={placeholder}
          type={show ? "text" : "password"}
          autoComplete={autoComplete}
          className="w-full rounded-2xl border border-[#dccabc] bg-[#fffdfa] px-4 py-4 pr-24 text-base font-semibold text-[#4e342e] outline-none transition placeholder:text-[#b09d90] focus:border-[#9c7257] focus:ring-4 focus:ring-[#ead8ca]"
        />

        <button
          type="button"
          onClick={() => setShow((prev) => !prev)}
          className="absolute right-3 top-1/2 -translate-y-1/2 rounded-full border border-[#d9c6b8] bg-white px-3 py-1 text-xs font-bold text-[#7b523a] transition hover:bg-[#f7efe8]"
        >
          {show ? "非表示" : "表示"}
        </button>
      </div>
    </label>
  );
}
