import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import { useI18n } from '../i18n'
import { BackupInfo } from '../types'
import {
  GetAdminerStatus,
  StartAdminer,
  StopAdminer,
  ListBackups,
  ListDatabases,
  CreateBackup,
  RestoreBackup,
  DeleteBackup,
} from '../../wailsjs/go/main/App'
import {
  Database as DatabaseIcon, AlertTriangle, Globe, Square,
  ExternalLink, HardDrive, Download, Upload, Trash2, Check, X,
} from 'lucide-react'

interface AdminerStatus {
  running:       boolean
  url:           string
  adminer_found: boolean
  php_found:     boolean
  php_path:      string
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

export default function Database() {
  const { services } = useServiceStore()
  const { t } = useI18n()
  const mysql = services.find(s => (s.name as string) === 'mysql')

  const [status, setStatus]       = useState<AdminerStatus | null>(null)
  const [starting, setStarting]   = useState(false)
  const [error, setError]         = useState('')

  // Backup state
  const [backups, setBackups]         = useState<BackupInfo[]>([])
  const [databases, setDatabases]     = useState<string[]>([])
  const [selectedDb, setSelectedDb]   = useState('all')
  const [backingUp, setBackingUp]     = useState(false)
  const [restoring, setRestoring]     = useState<string | null>(null)
  const [confirmRestore, setConfirmRestore] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete]   = useState<string | null>(null)
  const [backupSuccess, setBackupSuccess]   = useState('')

  const mysqlRunning = mysql?.status === 'running'

  const refresh = async () => {
    const s = await GetAdminerStatus()
    setStatus(s as AdminerStatus)
  }

  const refreshBackups = async () => {
    try {
      const list = await ListBackups()
      setBackups((list || []) as BackupInfo[])
    } catch { /* MySQL not running */ }
  }

  const refreshDatabases = async () => {
    try {
      const dbs = await ListDatabases()
      setDatabases(dbs || [])
    } catch { /* MySQL not running */ }
  }

  useEffect(() => {
    refresh()
    refreshBackups()
  }, [])

  useEffect(() => {
    if (mysqlRunning) refreshDatabases()
  }, [mysqlRunning])

  const handleStartAdminer = async () => {
    setStarting(true)
    setError('')
    try {
      await StartAdminer()
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

  const handleCreateBackup = async () => {
    setBackingUp(true)
    setError('')
    setBackupSuccess('')
    try {
      await CreateBackup(selectedDb === 'all' ? '' : selectedDb)
      await refreshBackups()
      setBackupSuccess(t.db_backup_success)
      setTimeout(() => setBackupSuccess(''), 3000)
    } catch (e: any) {
      setError(e?.toString() ?? 'Backup failed')
    } finally {
      setBackingUp(false)
    }
  }

  const handleRestore = async (filename: string) => {
    setConfirmRestore(null)
    setRestoring(filename)
    setError('')
    setBackupSuccess('')
    try {
      await RestoreBackup(filename)
      setBackupSuccess(t.db_restore_success)
      setTimeout(() => setBackupSuccess(''), 3000)
    } catch (e: any) {
      setError(e?.toString() ?? 'Restore failed')
    } finally {
      setRestoring(null)
    }
  }

  const handleDeleteBackup = async (filename: string) => {
    setConfirmDelete(null)
    try {
      await DeleteBackup(filename)
      await refreshBackups()
    } catch (e: any) {
      setError(e?.toString() ?? 'Delete failed')
    }
  }

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-white">{t.db_title}</h2>
        <p className="text-gray-400 text-sm mt-1">{t.db_desc}</p>
      </div>

      {error && (
        <div className="text-red-400 text-sm bg-red-500/10 rounded-lg px-4 py-3 flex items-center gap-2">
          <X size={14} className="flex-shrink-0" /> {error}
        </div>
      )}
      {backupSuccess && (
        <div className="text-green-400 text-sm bg-green-500/10 rounded-lg px-4 py-3 flex items-center gap-2">
          <Check size={14} className="flex-shrink-0" /> {backupSuccess}
        </div>
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
              <p className="text-xs text-gray-500 mt-1">{t.db_adminer_desc}</p>
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

        {status && !status.php_found && (
          <div className="mt-3 text-xs text-yellow-500 bg-yellow-500/10 rounded-lg px-3 py-2 flex items-center gap-1.5">
            <AlertTriangle size={14} className="text-yellow-400 flex-shrink-0" />
            <span>{t.db_php_not_found}</span>
          </div>
        )}
      </div>

      {/* Backups Card */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5">
        <div className="flex items-start justify-between gap-4 mb-4">
          <div className="flex items-start gap-3">
            <HardDrive size={20} className="text-gray-300 mt-0.5" />
            <div>
              <span className="text-white font-semibold">{t.db_backups}</span>
              <p className="text-xs text-gray-500 mt-1">{t.db_backups_desc}</p>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-shrink-0">
            <select
              value={selectedDb}
              onChange={e => setSelectedDb(e.target.value)}
              disabled={!mysqlRunning}
              className="bg-[#0f1420] border border-[#2a3347] rounded-lg px-2 py-1.5 text-xs text-gray-300 focus:outline-none focus:border-blue-500"
            >
              <option value="all">{t.db_backup_all}</option>
              {databases.map(db => (
                <option key={db} value={db}>{db}</option>
              ))}
            </select>
            <button
              onClick={handleCreateBackup}
              disabled={!mysqlRunning || backingUp}
              className="px-3 py-1.5 text-xs rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors disabled:opacity-40 flex items-center gap-1.5"
            >
              <Download size={12} />
              {backingUp ? t.db_creating_backup : t.db_create_backup}
            </button>
          </div>
        </div>

        {!mysqlRunning && (
          <div className="text-xs text-yellow-500 bg-yellow-500/10 rounded-lg px-3 py-2 flex items-center gap-1.5 mb-4">
            <AlertTriangle size={14} className="text-yellow-400 flex-shrink-0" />
            <span>{t.db_mysql_required}</span>
          </div>
        )}

        {/* Backup list */}
        {backups.length === 0 ? (
          <p className="text-gray-600 text-sm text-center py-6">{t.db_no_backups}</p>
        ) : (
          <div className="flex flex-col gap-2">
            {backups.map(bk => (
              <div
                key={bk.name}
                className="flex items-center gap-3 px-3 py-2 rounded-lg bg-[#0f1420] border border-transparent hover:border-[#2a3347] transition-colors"
              >
                <HardDrive size={14} className="text-gray-500 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-white font-mono truncate" title={bk.name}>{bk.name}</p>
                  <div className="flex gap-3 text-xs text-gray-500 mt-0.5">
                    <span>{bk.database}</span>
                    <span>{formatSize(bk.size)}</span>
                    <span>{new Date(bk.created_at).toLocaleString()}</span>
                  </div>
                </div>

                <div className="flex items-center gap-1.5 flex-shrink-0">
                  {confirmRestore === bk.name ? (
                    <>
                      <span className="text-xs text-yellow-400 mr-1">{t.db_confirm_restore}</span>
                      <button
                        onClick={() => handleRestore(bk.name)}
                        className="px-2 py-1 rounded-lg text-xs bg-yellow-500 text-black font-medium hover:bg-yellow-400 transition-colors"
                      >
                        {t.yes}
                      </button>
                      <button
                        onClick={() => setConfirmRestore(null)}
                        className="px-2 py-1 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                      >
                        {t.no}
                      </button>
                    </>
                  ) : confirmDelete === bk.name ? (
                    <>
                      <button
                        onClick={() => handleDeleteBackup(bk.name)}
                        className="px-2 py-1 rounded-lg text-xs bg-red-500 text-white hover:bg-red-600 transition-colors"
                      >
                        {t.yes}
                      </button>
                      <button
                        onClick={() => setConfirmDelete(null)}
                        className="px-2 py-1 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                      >
                        {t.no}
                      </button>
                    </>
                  ) : (
                    <>
                      <button
                        onClick={() => setConfirmRestore(bk.name)}
                        disabled={!mysqlRunning || restoring === bk.name}
                        className="px-2.5 py-1 rounded-lg text-xs bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors disabled:opacity-40 flex items-center gap-1"
                      >
                        <Upload size={12} />
                        {restoring === bk.name ? t.db_restoring : t.db_backup_restore}
                      </button>
                      <button
                        onClick={() => setConfirmDelete(bk.name)}
                        className="px-2.5 py-1 rounded-lg text-xs bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                      >
                        <Trash2 size={12} />
                      </button>
                    </>
                  )}
                </div>
              </div>
            ))}
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
