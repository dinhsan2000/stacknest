import { useState } from 'react'
import { KillConflictProcess, StartService } from '../../wailsjs/go/main/App'
import { useI18n, tt } from '../i18n'
import { Zap } from 'lucide-react'

interface ConflictInfo {
  port: number
  pid: number
  process: string
  in_use: boolean
}

interface Props {
  conflict: ConflictInfo
  serviceName: string
  onClose: () => void
  onResolved: () => void
}

export default function PortConflictModal({ conflict, serviceName, onClose, onResolved }: Props) {
  const [killing, setKilling] = useState(false)
  const [error, setError] = useState('')
  const { t } = useI18n()

  const handleKillAndStart = async () => {
    setKilling(true)
    setError('')
    try {
      await KillConflictProcess(conflict.pid)
      // Wait for OS to release port
      await new Promise(r => setTimeout(r, 800))
      await StartService(serviceName)
      onResolved()
    } catch (e: any) {
      setError(e?.toString() ?? 'Unknown error')
    } finally {
      setKilling(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#1e2535] border border-red-500/30 rounded-xl p-6 max-w-md w-full shadow-2xl">
        {/* Icon + Title */}
        <div className="flex items-start gap-4 mb-5">
          <span className="text-red-400"><Zap size={20} /></span>
          <div>
            <h3 className="text-white font-semibold text-lg">{t.pc_title}</h3>
            <p className="text-gray-400 text-sm mt-1">
              {tt(t.pc_port_in_use, { port: conflict.port })}{' '}
              {tt(t.pc_unable_start, { service: serviceName })}
            </p>
          </div>
        </div>

        {/* Conflict info */}
        <div className="bg-[#0f1420] rounded-lg p-4 mb-5 flex flex-col gap-2 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-400">{t.pc_process}</span>
            <span className="text-white font-mono">{conflict.process || 'unknown'}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-400">PID</span>
            <span className="text-white font-mono">{conflict.pid > 0 ? conflict.pid : '—'}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-400">Port</span>
            <span className="text-red-400 font-mono">{conflict.port}</span>
          </div>
        </div>

        {error && (
          <p className="text-red-400 text-xs bg-red-500/10 rounded p-2 mb-4">{error}</p>
        )}

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 py-2 rounded-lg bg-[#0f1420] text-gray-400 hover:text-white text-sm transition-colors"
          >
            {t.cancel}
          </button>

          {conflict.pid > 0 ? (
            <button
              onClick={handleKillAndStart}
              disabled={killing}
              className="flex-1 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 text-sm font-medium transition-colors disabled:opacity-50"
            >
              {killing ? t.pc_killing : tt(t.pc_kill_start, { service: serviceName })}
            </button>
          ) : (
            <button
              onClick={onClose}
              className="flex-1 py-2 rounded-lg bg-yellow-500/20 text-yellow-400 text-sm"
            >
              {tt(t.pc_free_port, { port: conflict.port })}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
