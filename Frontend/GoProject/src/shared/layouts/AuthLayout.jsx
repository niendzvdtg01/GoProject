export function AuthLayout({ children }) {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 p-4">
      <section className="grid w-full max-w-5xl overflow-hidden rounded-lg border border-slate-200 bg-white shadow-xl md:grid-cols-[1fr_430px]">
        <div className="flex min-h-80 flex-col justify-center border-b border-slate-200 bg-gradient-to-br from-slate-50 to-sky-50 p-8 md:border-b-0 md:border-r md:p-12">
          <span className="mb-3 text-xs font-extrabold uppercase text-sky-700">GoProject Gateway Client</span>
          <h1 className="text-3xl font-extrabold leading-tight text-slate-950">Workspace management</h1>
          <p className="mt-3 max-w-xl text-slate-600">
            Authentication, dashboard, and team management are organized for clarity and future growth.
          </p>
        </div>
        {children}
      </section>
    </main>
  )
}
