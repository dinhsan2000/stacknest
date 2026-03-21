import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import { useI18n } from '../i18n'
import {
  GetAdminerStatus,
  StartAdminer,
  StopAdminer,
} from '../../wailsjs/go/main/App'
import { Database as DatabaseIcon, AlertTriangle, Globe, Square, ExternalLink } from 'lucide-react'

interface AdminerStatus {
  running:       boolean
  url:           string
  adminer_found: boolean
  php_found:     boolean
  php_path:      string
}

export default function Database() {
  const { services } = useServiceStore()
  const { t } = useI18n()
  const mysql = services.find(s => (s.name as string) === 'mysql')

  const [status, setStatus]       = useState<AdminerStatus | null>(null)
  const [starting, setStarting]   = useState(false)
  const [error, setError]         = useState('')

  const refresh = async () => {
    const s = await GetAdminerStatus()
    setStatus(s as AdminerStatus)
  }

  useEffect(() => {
    refresh()
  }, [])

  const handleStartAdminer = async () => {
    setStarting(true)
    setError('')
    try {
      await StartAdminer()  // Go side opens browser automatically
      await refresh()
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to start Adminer')
    } finally {
      setStarting(false)
    }
  }

  const handleStopAdminer = async () => {
    await StopAdminer()
    await refresh()
  }

  const mysqlRunning = mysql?.status === 'running'

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-white">{t.db_title}</h2>
        <p className="text-gray-400 text-sm mt-1">{t.db_desc}</p>
      </div>

      {error && (
        <div className="text-red-400 text-sm bg-red-500/10 rounded-lg px-4 py-3">{error}</div>
      )}

      {/* MySQL Status Card */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
        <div className="flex items-center gap-3">
          <DatabaseIcon size={20} className="text-gray-300" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="text-white font-semibold">{t.db_mysql_label}</span>
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium
                ${mysqlRunning
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
                }`}
              >
                {mysqlRunning ? (
                  <span className="flex items-center gap-1">
                    <span className="inline-block w-2 h-2 rounded-full bg-green-400" />
                    {t.status_running}
                  </span>
                ) : (
                  <span className="flex items-center gap-1">
                    <span className="inline-block w-2 h-2 rounded-full border border-gray-400" />
                    {t.status_stopped}
                  </span>
                )}
              </span>
            </div>
            <div className="flex gap-4 mt-2 text-xs text-gray-500">
              <span>{t.db_host} <span className="text-gray-300 font-mono">127.0.0.1</span></span>
              <span>{t.db_port} <span className="text-gray-300 font-mono">{mysql?.port ?? 3306}</span></span>
              <span>{t.db_username} <span className="text-gray-300 font-mono">root</span></span>
              <span>{t.db_password} <span className="text-gray-500 italic">{t.db_password_empty}</span></span>
            </div>
          </div>
        </div>

        {!mysqlRunning && (
          <p className="text-yellow-500 text-xs mt-3 flex items-center gap-1.5">
            <AlertTriangle size={14} className="text-yellow-400" />
            <span>{t.db_mysql_start_warn}</span>
          </p>
        )}
      </div>

      {/* Adminer Card */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <Globe size={20} className="text-gray-300 mt-0.5" />
            <div>
              <div className="flex items-center gap-2">
                <span className="text-white font-semibold">{t.db_adminer}</span>
                <span className="text-xs text-gray-500 bg-[#0f1420] px-2 py-0.5 rounded">{t.db_web_based}</span>
                {status?.running && (
                  <span className="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full flex items-center gap-1">
                    <span className="inline-block w-2 h-2 rounded-full bg-green-400" />
                    {t.db_adminer_running}
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-500 mt-1">
                {t.db_adminer_desc}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-shrink-0">
            {status?.running ? (
              <>
                <a
                  href={status.url}
                  onClick={e => { e.preventDefault(); StartAdminer() }}
                  className="px-4 py-2 text-sm rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors flex items-center gap-1.5"
                >
                  {t.db_open} <ExternalLink size={14} />
                </a>
                <button
                  onClick={handleStopAdminer}
                  className="px-3 py-2 text-sm rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
                  title={t.db_stop_adminer}
                >
                  <Square size={12} />
                </button>
              </>
            ) : (
              <button
                onClick={handleStartAdminer}
                disabled={starting || !status?.adminer_found || !status?.php_found}
                className="px-4 py-2 text-sm rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors disabled:opacity-40"
              >
                {starting ? t.db_starting : t.db_start_open}
              </button>
            )}
          </div>
        </div>

        {/* Missing PHP */}
        {status && !status.php_found && (
          <div className="mt-3 text-xs text-yellow-500 bg-yellow-500/10 rounded-lg px-3 py-2 flex items-center gap-1.5">
            <AlertTriangle size={14} className="text-yellow-400 flex-shrink-0" />
            <span>{t.db_php_not_found}</span>
          </div>
        )}
      </div>

      {/* Info box */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-4 text-xs text-gray-400">
        <p className="font-semibold text-gray-300 mb-2">{t.db_credentials_title}</p>
        <div className="grid grid-cols-2 gap-1 font-mono">
          <span className="text-gray-500">{t.db_server}</span>   <span>127.0.0.1</span>
          <span className="text-gray-500">{t.db_port}</span>     <span>3306</span>
          <span className="text-gray-500">{t.db_username}</span> <span>root</span>
          <span className="text-gray-500">{t.db_password}</span> <span className="text-gray-600 italic">{t.db_password_empty}</span>
        </div>
      </div>
    </div>
  )
}
