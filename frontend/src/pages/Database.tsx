import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import {
  GetAdminerStatus,
  StartAdminer,
  StopAdminer,
  OpenHeidiSQL,
} from '../../wailsjs/go/main/App'

interface AdminerStatus {
  running:       boolean
  url:           string
  adminer_found: boolean
  adminer_path:  string
  php_found:     boolean
  php_path:      string
  heidisql_path: string
}

export default function Database() {
  const { services } = useServiceStore()
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

  const handleOpenHeidiSQL = async () => {
    try {
      await OpenHeidiSQL()
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to open HeidiSQL')
    }
  }

  const mysqlRunning = mysql?.status === 'running'

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-white">Database Manager</h2>
        <p className="text-gray-400 text-sm mt-1">Launch web-based or native database management tools</p>
      </div>

      {error && (
        <div className="text-red-400 text-sm bg-red-500/10 rounded-lg px-4 py-3">{error}</div>
      )}

      {/* MySQL Status Card */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
        <div className="flex items-center gap-3">
          <span className="text-2xl">🗄</span>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="text-white font-semibold">MySQL</span>
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium
                ${mysqlRunning
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
                }`}
              >
                {mysqlRunning ? '● Running' : '○ Stopped'}
              </span>
            </div>
            <div className="flex gap-4 mt-2 text-xs text-gray-500">
              <span>Host: <span className="text-gray-300 font-mono">127.0.0.1</span></span>
              <span>Port: <span className="text-gray-300 font-mono">{mysql?.port ?? 3306}</span></span>
              <span>User: <span className="text-gray-300 font-mono">root</span></span>
              <span>Password: <span className="text-gray-500 italic">none</span></span>
            </div>
          </div>
        </div>

        {!mysqlRunning && (
          <p className="text-yellow-500 text-xs mt-3 flex items-center gap-1.5">
            <span>⚠</span>
            <span>Start MySQL from the Services page before opening database tools</span>
          </p>
        )}
      </div>

      {/* Adminer Card */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <span className="text-2xl mt-0.5">🌐</span>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-white font-semibold">Adminer</span>
                <span className="text-xs text-gray-500 bg-[#0f1420] px-2 py-0.5 rounded">web-based</span>
                {status?.running && (
                  <span className="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">
                    ● Running
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-500 mt-1">
                Lightweight single-file PHP database manager
              </p>
              {status?.adminer_path && (
                <p className="text-xs text-gray-600 mt-0.5 font-mono truncate" title={status.adminer_path}>
                  {status.adminer_path}
                </p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2 flex-shrink-0">
            {status?.running ? (
              <>
                <a
                  href={status.url}
                  onClick={e => { e.preventDefault(); StartAdminer() }}
                  className="px-4 py-2 text-sm rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors"
                >
                  Open ↗
                </a>
                <button
                  onClick={handleStopAdminer}
                  className="px-3 py-2 text-sm rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
                  title="Stop Adminer server"
                >
                  ■
                </button>
              </>
            ) : (
              <button
                onClick={handleStartAdminer}
                disabled={starting || !status?.adminer_found || !status?.php_found}
                className="px-4 py-2 text-sm rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors disabled:opacity-40"
              >
                {starting ? 'Starting...' : 'Start & Open'}
              </button>
            )}
          </div>
        </div>

        {/* Missing dependencies */}
        {status && !status.adminer_found && (
          <div className="mt-3 text-xs text-yellow-500 bg-yellow-500/10 rounded-lg px-3 py-2">
            ⚠ Adminer not found. Expected at{' '}
            <span className="font-mono">C:\laragon\etc\apps\adminer\index.php</span>
          </div>
        )}
        {status && !status.php_found && (
          <div className="mt-3 text-xs text-yellow-500 bg-yellow-500/10 rounded-lg px-3 py-2">
            ⚠ PHP not found. Configure a PHP version in the PHP Versions page.
          </div>
        )}
      </div>

      {/* HeidiSQL Card */}
      {status?.heidisql_path && (
        <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl mt-0.5">🖥</span>
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-white font-semibold">HeidiSQL</span>
                  <span className="text-xs text-gray-500 bg-[#0f1420] px-2 py-0.5 rounded">native client</span>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  Powerful native GUI for MySQL, MariaDB, PostgreSQL
                </p>
                <p className="text-xs text-gray-600 mt-0.5 font-mono truncate" title={status.heidisql_path}>
                  {status.heidisql_path}
                </p>
              </div>
            </div>
            <button
              onClick={handleOpenHeidiSQL}
              className="flex-shrink-0 px-4 py-2 text-sm rounded-lg bg-[#2a3347] text-gray-300 hover:text-white hover:bg-[#3a4357] transition-colors"
            >
              Open HeidiSQL
            </button>
          </div>
        </div>
      )}

      {/* Info box */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-4 text-xs text-gray-400">
        <p className="font-semibold text-gray-300 mb-2">Default credentials</p>
        <div className="grid grid-cols-2 gap-1 font-mono">
          <span className="text-gray-500">Server:</span>   <span>127.0.0.1</span>
          <span className="text-gray-500">Port:</span>     <span>3306</span>
          <span className="text-gray-500">Username:</span> <span>root</span>
          <span className="text-gray-500">Password:</span> <span className="text-gray-600 italic">leave empty</span>
        </div>
      </div>
    </div>
  )
}
