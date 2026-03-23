import { useEffect, useState, useCallback } from 'react'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useI18n, tt } from '../i18n'
import { AlertTriangle, CheckCircle, XCircle, Info, X } from 'lucide-react'

type ToastType = 'success' | 'error' | 'warning' | 'info'

interface Toast {
  id: number
  type: ToastType
  message: string
}

let nextId = 0

const ICONS: Record<ToastType, React.ReactNode> = {
  success: <CheckCircle size={16} className="text-green-400 flex-shrink-0" />,
  error:   <XCircle size={16} className="text-red-400 flex-shrink-0" />,
  warning: <AlertTriangle size={16} className="text-yellow-400 flex-shrink-0" />,
  info:    <Info size={16} className="text-blue-400 flex-shrink-0" />,
}

const BG: Record<ToastType, string> = {
  success: 'bg-green-500/10 border-green-500/30',
  error:   'bg-red-500/10 border-red-500/30',
  warning: 'bg-yellow-500/10 border-yellow-500/30',
  info:    'bg-blue-500/10 border-blue-500/30',
}

export default function ToastContainer() {
  const { t } = useI18n()
  const [toasts, setToasts] = useState<Toast[]>([])

  const addToast = useCallback((type: ToastType, message: string) => {
    const id = ++nextId
    setToasts(prev => [...prev, { id, type, message }])
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id))
    }, 5000)
  }, [])

  const removeToast = useCallback((id: number) => {
    setToasts(prev => prev.filter(t => t.id !== id))
  }, [])

  useEffect(() => {
    // Prevent duplicate
    if ((window as any).__stacknest_toast_init) return
    ;(window as any).__stacknest_toast_init = true

    EventsOn('service:crashed', (data: { name: string; error: string; recover: boolean }) => {
      if (data.recover) {
        addToast('warning', tt(t.notif_service_recovered, { service: data.name }))
      } else {
        addToast('error', tt(t.notif_service_crashed, { service: data.name, error: data.error }))
      }
    })

    EventsOn('service:crash-loop', (data: { name: string }) => {
      addToast('error', tt(t.notif_crash_loop, { service: data.name }))
    })

    EventsOn('binary:done', (data: { service: string; version: string; error: string }) => {
      if (data.error) {
        addToast('error', tt(t.notif_download_failed, { service: data.service, version: data.version }))
      } else {
        addToast('success', tt(t.notif_download_done, { service: data.service, version: data.version }))
      }
    })
  }, [addToast, t])

  if (toasts.length === 0) return null

  return (
    <div className="fixed top-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
      {toasts.map(toast => (
        <div
          key={toast.id}
          className={`flex items-start gap-2 px-4 py-3 rounded-xl border shadow-lg backdrop-blur-sm text-sm animate-slide-in ${BG[toast.type]}`}
        >
          {ICONS[toast.type]}
          <span className="text-gray-200 flex-1">{toast.message}</span>
          <button
            onClick={() => removeToast(toast.id)}
            className="text-gray-500 hover:text-gray-300 flex-shrink-0"
          >
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  )
}
