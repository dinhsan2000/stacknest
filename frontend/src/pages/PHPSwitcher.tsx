import { useEffect, useState } from 'react'
import { GetPHPInstalls, SwitchPHP, AddPHPPath, SelectFolder } from '../../wailsjs/go/main/App'

interface PHPInstall {
  version: string
  major:   string
  path:    string
  active:  boolean
}

const versionColor = (major: string) => {
  const v = parseFloat(major)
  if (v >= 8.2) return 'text-green-400'
  if (v >= 8.0) return 'text-blue-400'
  if (v >= 7.4) return 'text-yellow-400'
  return 'text-red-400'
}

const versionBg = (major: string) => {
  const v = parseFloat(major)
  if (v >= 8.2) return 'bg-green-500/10 border-green-500/30'
  if (v >= 8.0) return 'bg-blue-500/10 border-blue-500/30'
  if (v >= 7.4) return 'bg-yellow-500/10 border-yellow-500/30'
  return 'bg-red-500/10 border-red-500/30'
}

export default function PHPSwitcher() {
  const [installs, setInstalls] = useState<PHPInstall[]>([])
  const [loading,  setLoading]  = useState(true)
  const [switching, setSwitching] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const refresh = async () => {
    setLoading(true)
    try {
      const data = await GetPHPInstalls()
      setInstalls((data || []) as PHPInstall[])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { refresh() }, [])

  const handleSwitch = async (php: PHPInstall) => {
    if (php.active) return
    setSwitching(php.path)
    setError('')
    setSuccess('')
    try {
      await SwitchPHP(php.path)
      setSuccess(`Switched to PHP ${php.version}`)
      await refresh()
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to switch PHP version')
    } finally {
      setSwitching('')
    }
  }

  const handleAddPath = async () => {
    const dir = await SelectFolder()
    if (!dir) return
    try {
      await AddPHPPath(dir)
      await refresh()
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to add path')
    }
  }

  const active = installs.find(p => p.active)

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">PHP Version Switcher</h2>
          <p className="text-gray-400 text-sm mt-1">
            {active
              ? <>Active: <span className="text-green-400 font-mono">PHP {active.version}</span></>
              : 'No PHP found on this machine'}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleAddPath}
            className="px-3 py-1.5 text-xs rounded-lg bg-[#1e2535] text-gray-400 hover:text-white transition-colors"
          >
            + Add Path
          </button>
          <button
            onClick={refresh}
            disabled={loading}
            className="px-3 py-1.5 text-xs rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors disabled:opacity-50"
          >
            {loading ? '⟳ Scanning...' : '⟳ Rescan'}
          </button>
        </div>
      </div>

      {/* Alerts */}
      {error   && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg p-3">{error}</p>}
      {success && <p className="text-green-400 text-sm bg-green-500/10 rounded-lg p-3">✓ {success}</p>}

      {/* Loading state */}
      {loading && installs.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <p className="text-3xl mb-3">🔍</p>
          <p>Scanning for PHP installations...</p>
        </div>
      )}

      {/* No PHP found */}
      {!loading && installs.length === 0 && (
        <div className="text-center py-12 bg-[#1e2535] rounded-xl border border-[#2a3347]">
          <p className="text-3xl mb-3">🐘</p>
          <p className="text-gray-300 font-medium">No PHP found</p>
          <p className="text-gray-500 text-sm mt-2 mb-4">
            Install PHP or click "+ Add Path" to point to a custom directory
          </p>
          <button
            onClick={handleAddPath}
            className="px-4 py-2 rounded-lg bg-blue-500 text-white text-sm hover:bg-blue-600 transition-colors"
          >
            Add Custom PHP Path
          </button>
        </div>
      )}

      {/* Version cards */}
      <div className="flex flex-col gap-3">
        {installs.map(php => (
          <div
            key={php.path}
            className={`border rounded-xl p-4 flex items-center gap-4 transition-all
              ${php.active
                ? `${versionBg(php.major)} ring-1 ring-green-500/40`
                : 'bg-[#1e2535] border-[#2a3347] hover:border-[#3a4357]'
              }`}
          >
            {/* PHP elephant icon + version badge */}
            <div className="flex flex-col items-center gap-1 w-16">
              <span className="text-3xl">🐘</span>
              <span className={`text-xs font-bold font-mono ${versionColor(php.major)}`}>
                {php.major}
              </span>
            </div>

            {/* Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-white font-semibold font-mono">PHP {php.version}</span>
                {php.active && (
                  <span className="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">
                    Active
                  </span>
                )}
              </div>
              <p className="text-gray-500 text-xs mt-1 truncate" title={php.path}>
                {php.path}
              </p>
            </div>

            {/* Switch button */}
            <div>
              {php.active ? (
                <span className="text-xs text-green-400">✓ In use</span>
              ) : (
                <button
                  onClick={() => handleSwitch(php)}
                  disabled={!!switching}
                  className="px-4 py-2 rounded-lg bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-colors disabled:opacity-50 whitespace-nowrap"
                >
                  {switching === php.path ? 'Switching...' : 'Use this version'}
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Info box */}
      {installs.length > 0 && (
        <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-4 text-xs text-gray-400">
          <p className="font-semibold text-gray-300 mb-2">How switching works</p>
          <ul className="flex flex-col gap-1 list-disc list-inside">
            <li>Selected PHP path is saved to Stacknest config</li>
            <li>If PHP-FPM service is running, it will be restarted automatically</li>
            <li>New terminal sessions will use the active PHP version</li>
            <li>On Windows, a <code className="bg-[#0f1420] px-1 rounded">current</code> symlink is updated in Laragon's PHP folder</li>
          </ul>
        </div>
      )}
    </div>
  )
}
