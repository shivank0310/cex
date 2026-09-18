import clsx from "clsx";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "buy" | "sell" | "ghost";
  size?: "sm" | "md" | "lg";
}

export function Button({
  variant = "primary",
  size = "md",
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={clsx(
        "inline-flex items-center justify-center rounded-xl font-semibold transition-all duration-200",
        "disabled:opacity-50 disabled:cursor-not-allowed",
        size === "sm" && "px-3 py-1.5 text-sm",
        size === "md" && "px-5 py-2.5 text-sm",
        size === "lg" && "px-8 py-3.5 text-base",
        variant === "primary" &&
          "bg-gradient-to-r from-violet-600 to-cyan-500 text-white hover:shadow-lg hover:shadow-violet-500/30",
        variant === "secondary" &&
          "bg-white/10 text-white border border-white/20 hover:bg-white/15",
        variant === "buy" &&
          "bg-emerald-500 text-white hover:bg-emerald-400",
        variant === "sell" &&
          "bg-rose-500 text-white hover:bg-rose-400",
        variant === "ghost" && "text-slate-300 hover:text-white hover:bg-white/5",
        className
      )}
      {...props}
    >
      {children}
    </button>
  );
}
