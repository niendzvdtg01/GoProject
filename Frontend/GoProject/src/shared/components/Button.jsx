export function Button({
  children,
  className = '',
  isLoading = false,
  variant = 'primary',
  ...props
}) {
  const variants = {
    primary: 'bg-sky-700 text-white hover:bg-sky-800 border-transparent',
    secondary: 'bg-white text-slate-900 hover:bg-slate-50 border-slate-300',
    ghost: 'bg-transparent text-slate-700 hover:bg-slate-100 border-transparent',
    danger: 'bg-red-600 text-white hover:bg-red-700 border-transparent',
  }

  return (
    <button
      className={`inline-flex min-h-10 items-center justify-center rounded-md border px-4 py-2 text-sm font-bold transition ${variants[variant]} ${className}`}
      disabled={isLoading || props.disabled}
      {...props}
    >
      {isLoading ? 'Đang xử lý' : children}
    </button>
  )
}
