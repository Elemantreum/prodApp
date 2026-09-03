import type { ButtonHTMLAttributes } from "react";

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger";
  loading?: boolean;
}

export function Button({ variant = "primary", loading = false, disabled, children, className, ...rest }: Props) {
  return (
    <button
      className={["btn", `btn--${variant}`, className].filter(Boolean).join(" ")}
      disabled={disabled || loading}
      {...rest}
    >
      {loading ? <span className="spinner spinner--inline" /> : children}
    </button>
  );
}
