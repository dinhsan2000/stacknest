import { useEffect, useState, useCallback } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { oneDark } from '@codemirror/theme-one-dark'
import { StreamLanguage } from '@codemirror/language'
import { nginx } from '@codemirror/legacy-modes/mode/nginx'
import { properties } from '@codemirror/legacy-modes/mode/properties'
import {
  GetServiceConfigs,
  ReadConfigFile,
  SaveConfigFile,
  GetConfigBackups,
  RestoreConfigBackup,
  RestartService,
} from '../../wailsjs/go/main/App'
import { Code2, Database, Layers, Server, Zap } from 'lucide-react'
import { ServiceIcon } from '../components/ServiceIcon'

interface ConfigFile {
  service:  string
  label:    string
  path:     string
  lang:     string
  writable: boolean
}

interface BackupInfo {
  path:       string
  created_at: string
  size_bytes: number
}

const SERVICES = [
  { id: 'apache', label: 'Apache',    icon: '🌐' },
  { id: 'nginx',  label: 'Nginx',     icon: '⚡' },
  { id: 'mysql',  label: 'MySQL',     icon: '🗄' },
  { id: 'php',    label: 'PHP',       icon: '🐘' },
]

function getExtensions(lang: string) {
  switch (lang) {
    case 'nginx':  return [StreamLanguage.define(nginx)]
    case 'ini':    return [StreamLanguage.define(properties)]
    case 'apache': return [StreamLanguage.define(nginx)] // closest available
    default:       return []
  }
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  return `${(bytes / 1024).toFixed(1)} KB`
}

export default function ConfigEditor() {
  const [service, setService]         = useState('apache')
  const [configs, setConfigs]         = useState<ConfigFile[]>([])
  const [selected, setSelected]       = useState<ConfigFile | null>(null)
  const [content, setContent]         = useState('')
  const [savedContent, setSavedContent] = useState('')
  const [backups, setBackups]         = useState<BackupInfo[]>([])
  const [showBackups, setShowBackups] = useState(false)
  const [loading, setLoading]           = useState(false)
  const [saving, setSaving]             = useState(false)
  const [restarting, setRestarting]     = useState(false)
  const [error, setError]               = useState('')
  const [success, setSuccess]           = useState('')

  const isDirty = content !== savedContent

  // Load config files for selected service
  useEffect(() => {
    setSelected(null)
    setContent('')
    setSavedContent('')
    setError('')
    setSuccess('')
    setShowBackups(false)
    GetServiceConfigs(service).then(data => {
      const files = (data || []) as ConfigFile[]
      setConfigs(files)
      if (files.length > 0) selectFile(files[0])
    })
  }, [service])

  const selectFile = useCallback(async (file: ConfigFile) => {
    if (isDirty && selected) {
      const ok = window.confirm('You have unsaved changes. Discard them?')
      if (!ok) return
    }
    setSelected(file)
    setError('')
    setSuccess('')
    setShowBackups(false)
    setLoading(true)
    try {
      const text = await ReadConfigFile(file.path)
      setContent(text)
      setSavedContent(text)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to read file')
    } finally {
      setLoading(false)
    }
  }, [isDirty, selected])

  const handleSave = async () => {
    if (!selected || !selected.writable) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      await SaveConfigFile(selected.path, content)
      setSavedContent(content)
      setSuccess('Saved successfully')
      const bups = await GetConfigBackups(selected.path)
      setBackups((bups || []) as BackupInfo[])
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveAndRestart = async () => {
    if (!selected || !selected.writable) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      await SaveConfigFile(selected.path, content)
      setSavedContent(content)
      const bups = await GetConfigBackups(selected.path)
      setBackups((bups || []) as BackupInfo[])
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to save')
      setSaving(false)
      return
    }
    setSaving(false)
    setRestarting(true)
    try {
      await RestartService(service)
      setSuccess(`Saved and restarted ${service}`)
    } catch (e: any) {
      setError(e?.toString() ?? `Failed to restart ${service}`)
    } finally {
      setRestarting(false)
    }
  }

  const handleShowBackups = async () => {
    if (!selected) return
    const bups = await GetConfigBackups(selected.path)
    setBackups((bups || []) as BackupInfo[])
    setShowBackups(v => !v)
  }

  const handleRestore = async (backup: BackupInfo) => {
    if (!selected) return
    const ok = window.confirm(`Restore backup from ${backup.created_at}?\nThis will overwrite the current file.`)
    if (!ok) return
    try {
      await RestoreConfigBackup(backup.path, selected.path)
      // Reload
      const text = await ReadConfigFile(selected.path)
      setContent(text)
      setSavedContent(text)
      setSuccess(`Restored backup from ${backup.created_at}`)
      setShowBackups(false)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to restore backup')
    }
  }

  return (
    <div className="flex flex-col gap-0 h-full">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div>
          <h2 className="text-2xl font-bold text-white">Config Editor</h2>
          <p className="text-gray-400 text-sm mt-1">Edit service configuration files with automatic backup</p>
        </div>
      </div>

      {/* Service tabs */}
      <div className="flex gap-1 mb-4 flex-shrink-0">
        {SERVICES.map(s => (
          <button
            key={s.id}
            onClick={() => setService(s.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors
              ${service === s.id
                ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                : 'bg-[#1e2535] text-gray-400 hover:text-white border border-transparent'
              }`}
          >
            <span><ServiceIcon name={s.id} /></span>
            <span>{s.label}</span>
          </button>
        ))}
      </div>

      {/* Main area */}
      <div className="flex gap-4 flex-1 min-h-0">
        {/* File list */}
        <div className="w-56 flex-shrink-0 flex flex-col gap-1">
          <p className="text-xs text-gray-500 uppercase tracking-wider mb-2 px-1">Config Files</p>
          {configs.length === 0 ? (
            <div className="text-center py-8 text-gray-600 text-sm">
              <p className="text-2xl mb-2">📭</p>
              <p>No config files found</p>
            </div>
          ) : (
            configs.map(f => (
              <button
                key={f.path}
                onClick={() => selectFile(f)}
                title={f.path}
                className={`text-left px-3 py-2.5 rounded-lg text-sm transition-colors
                  ${selected?.path === f.path
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                    : 'bg-[#1e2535] text-gray-400 hover:text-white border border-transparent'
                  }`}
              >
                <div className="font-medium truncate">{f.label}</div>
                {!f.writable && (
                  <div className="text-xs text-yellow-500 mt-0.5">read-only</div>
                )}
              </button>
            ))
          )}
        </div>

        {/* Editor panel */}
        <div className="flex-1 flex flex-col min-w-0 min-h-0">
          {/* Editor toolbar */}
          {selected && (
            <div className="flex items-center justify-between mb-2 flex-shrink-0">
              <div className="flex items-center gap-2 min-w-0">
                <span className="text-gray-500 text-xs truncate" title={selected.path}>
                  {selected.path}
                </span>
                {isDirty && (
                  <span className="text-xs text-yellow-400 bg-yellow-500/10 px-2 py-0.5 rounded-full flex-shrink-0">
                    unsaved
                  </span>
                )}
              </div>
              <div className="flex items-center gap-2 flex-shrink-0">
                <button
                  onClick={handleShowBackups}
                  className="px-3 py-1.5 text-xs rounded-lg bg-[#1e2535] text-gray-400 hover:text-white transition-colors"
                >
                  {showBackups ? 'Hide Backups' : 'Backups'}
                </button>
                <button
                  onClick={handleSave}
                  disabled={saving || restarting || !isDirty || !selected.writable}
                  className="px-4 py-1.5 text-xs rounded-lg bg-[#1e2535] border border-[#2a3347] text-gray-300 hover:text-white font-medium transition-colors disabled:opacity-40"
                >
                  {saving ? 'Saving...' : 'Save'}
                </button>
                <button
                  onClick={handleSaveAndRestart}
                  disabled={saving || restarting || !isDirty || !selected.writable}
                  className="px-4 py-1.5 text-xs rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors disabled:opacity-40"
                >
                  {restarting ? 'Restarting...' : saving ? 'Saving...' : 'Save & Restart'}
                </button>
              </div>
            </div>
          )}

          {/* Alerts */}
          {error && (
            <div className="mb-2 text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2 flex-shrink-0">
              {error}
            </div>
          )}
          {success && (
            <div className="mb-2 text-green-400 text-sm bg-green-500/10 rounded-lg px-3 py-2 flex-shrink-0">
              ✓ {success}
            </div>
          )}

          {/* Backup list */}
          {showBackups && backups.length > 0 && (
            <div className="mb-2 bg-[#1e2535] border border-[#2a3347] rounded-lg p-3 flex-shrink-0">
              <p className="text-xs text-gray-400 font-semibold mb-2">Available Backups</p>
              <div className="flex flex-col gap-1 max-h-36 overflow-y-auto">
                {backups.map((b, i) => (
                  <div key={i} className="flex items-center justify-between text-xs py-1 border-b border-[#2a3347] last:border-0">
                    <span className="text-gray-300 font-mono">{b.created_at}</span>
                    <div className="flex items-center gap-3">
                      <span className="text-gray-500">{formatBytes(b.size_bytes)}</span>
                      <button
                        onClick={() => handleRestore(b)}
                        className="text-blue-400 hover:text-blue-300 transition-colors"
                      >
                        Restore
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
          {showBackups && backups.length === 0 && (
            <div className="mb-2 text-gray-500 text-xs bg-[#1e2535] rounded-lg px-3 py-2 flex-shrink-0">
              No backups yet. Backups are created automatically when you save.
            </div>
          )}

          {/* CodeMirror editor */}
          {loading ? (
            <div className="flex-1 flex items-center justify-center text-gray-500">
              <p>Loading...</p>
            </div>
          ) : selected ? (
            <div className="flex-1 min-h-0 rounded-xl overflow-hidden border border-[#2a3347]">
              <CodeMirror
                value={content}
                onChange={setContent}
                theme={oneDark}
                extensions={getExtensions(selected.lang)}
                readOnly={!selected.writable}
                height="100%"
                style={{ height: '100%' }}
                basicSetup={{
                  lineNumbers: true,
                  foldGutter: true,
                  highlightActiveLine: true,
                  highlightSelectionMatches: true,
                }}
              />
            </div>
          ) : (
            <div className="flex-1 flex items-center justify-center text-gray-600">
              <div className="text-center">
                <p className="text-3xl mb-3">📝</p>
                <p>Select a config file to edit</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
