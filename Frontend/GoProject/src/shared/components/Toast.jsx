import { useEffect } from 'react'

export function Toast({ message, type = 'success', onClose, duration = 4000 }) {
  useEffect(() => {
    if (!message) return
    const timer = setTimeout(onClose, duration)
    return () => clearTimeout(timer)
  }, [message, duration, onClose])

  if (!message) return null

  const styles = {
    success: 'bg-emerald-600 text-white',
    error: 'bg-red-600 text-white',
    warning: 'bg-amber-500 text-white',
    info: 'bg-sky-600 text-white',
  }

  return (
    <div className={`fixed top-5 right-5 z-50 flex items-start gap-3 rounded-lg px-4 py-3 shadow-lg max-w-sm ${styles[type]}`}>
      <span className="text-sm font-medium leading-snug flex-1">{message}</span>
      <button
        onClick={onClose}
        className="shrink-0 opacity-75 hover:opacity-100 text-lg leading-none -mt-0.5"
      >
        ×
      </button>
    </div>
  )
}
