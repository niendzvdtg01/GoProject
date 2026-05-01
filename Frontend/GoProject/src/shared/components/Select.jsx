export function Select({ label, error, children, ...props }) {
  const id = props.id ?? props.name

  return (
    <label className="grid gap-1.5 text-sm font-bold text-slate-800" htmlFor={id}>
      {label}
      <select
        id={id}
        className="min-h-11 rounded-md border border-slate-300 bg-white px-3 py-2 font-normal text-slate-950 outline-none transition focus:border-sky-700 focus:ring-4 focus:ring-sky-100"
        aria-invalid={Boolean(error)}
        {...props}
      >
        {children}
      </select>
      {error ? <span className="text-xs font-semibold text-red-600">{error}</span> : null}
    </label>
  )
}
