import React, { useEffect, useState } from 'react'
import { Server, Database, Code2, Zap, Layers, RefreshCw, AlertTriangle, X, Check } from 'lucide-react'
import { useServiceStore } from '../store/serviceStore'
import { ReloadCatalog } from '../../wailsjs/go/main/App'
import { ServiceIcon } from '../components/ServiceIcon';
import { useI18n, tt } from '../i18n'

const SERVICE_META: Record<string, { label: string; icon: React.ReactNode }> = {
  apache: { label: 'Apache', icon: <Server size={16} className="text-gray-400" /> },
  nginx: { label: 'Nginx', icon: <Zap size={16} className="text-gray-400" /> },
  mysql: { label: 'MySQL', icon: <Database size={16} className="text-gray-400" /> },
  php: { label: 'PHP', icon: <Code2 size={16} className="text-gray-400" /> },
  redis: { label: 'Redis', icon: <Layers size={16} className="text-gray-400" /> },
}

export default function Binaries() {
  const {
    binaryStatus, downloadProgress, downloadErrors,
    fetchBinaryStatus, downloadBinary, cancelDownload, deleteBinary,
    setActiveVersion, dismissDownloadError,
  } = useServiceStore()

  const { t } = useI18n()
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null) // "service@version"
  const [deleteError, setDeleteError] = useState('')
  const [reloading, setReloading] = useState(false)

  useEffect(() => { fetchBinaryStatus() }, [])

  const handleReloadCatalog = async () => {
    setReloading(true)
    try {
      await ReloadCatalog()
      await fetchBinaryStatus()
    } finally {
      setReloading(false)
    }
  }

  const missingCount = binaryStatus.filter(s => !s.versions.some(v => v.installed)).length

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">{t.bin_title}</h2>
          <p className="text-gray-400 text-sm mt-1">
            {t.bin_desc}
            {' '}Binaries are saved at <code className="text-blue-400 text-xs">bin/{'{service}/{version}/'}</code>
          </p>
        </div>
        <button
          onClick={handleReloadCatalog}
          disabled={reloading}
          title={t.bin_reload_tooltip}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e2535] text-gray-400 hover:bg-[#2a3347] hover:text-white transition-colors disabled:opacity-50"
        >
          <RefreshCw size={14} className={reloading ? 'animate-spin' : ''} />
          {reloading ? t.bin_reloading : t.bin_reload}
        </button>
      </div>

      {missingCount > 0 && (
        <div className="flex items-center gap-3 bg-yellow-500/10 border border-yellow-500/20 rounded-xl px-4 py-3">
          <AlertTriangle size={16} className="text-yellow-400 mt-0.5 flex-shrink-0" />
          <p className="text-sm text-yellow-300">
            {tt(t.bin_missing_warn, { count: missingCount, s: missingCount > 1 ? 's' : '' })}
          </p>
        </div>
      )}

      {/* Global download errors */}
      {Object.entries(downloadErrors).map(([key, errMsg]) => (
        <div key={key} className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3">
          <X size={16} className="text-red-400 flex-shrink-0" />
          <p className="text-sm text-red-300 flex-1">
            <strong>{key.replace('@', ' v')}</strong>: {errMsg}
          </p>
          <button
            onClick={() => dismissDownloadError(key)}
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            {t.dismiss}
          </button>
        </div>
      ))}

      {deleteError && (
        <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3">
          <X size={16} className="text-red-400 flex-shrink-0" />
          <p className="text-sm text-red-300 flex-1">{deleteError}</p>
          <button
            onClick={() => setDeleteError('')}
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            {t.dismiss}
          </button>
        </div>
      )}

      <div className="flex flex-col gap-4">
        {binaryStatus.map(svc => {
          const meta = SERVICE_META[svc.service]
          const hasAnyInstalled = svc.versions.some(v => v.installed)
          const activeVersion = svc.versions.find(v => v.active)

          return (
            <div
              key={svc.service}
              className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-4"
            >
              {/* Service header */}
              <div className="flex items-center gap-3 mb-3">
                <span className="text-lg"><ServiceIcon name={svc.service} /></span>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-white font-semibold">{meta?.label ?? svc.service}</span>
                    {hasAnyInstalled ? (
                      <span className="text-xs px-2 py-0.5 rounded-full bg-green-500/15 text-green-400">
                        {activeVersion ? `${t.bin_active}: v${activeVersion.version}` : t.bin_installed}
                      </span>
                    ) : (
                      <span className="text-xs px-2 py-0.5 rounded-full bg-gray-500/15 text-gray-500">
                        {t.bin_not_installed}
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {/* Version list */}
              <div className="flex flex-col gap-2">
                {svc.versions.map(ver => {
                  const key = `${svc.service}@${ver.version}`
                  const progress = downloadProgress[key]
                  const isDownloading = progress !== undefined

                  return (
                    <div
                      key={ver.version}
                      className={`flex items-center gap-3 px-3 py-2 rounded-lg ${ver.active
                          ? 'bg-blue-500/10 border border-blue-500/20'
                          : 'bg-[#0f1420] border border-transparent'
                        }`}
                    >
                      {/* Version indicator dot */}
                      <div className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${isDownloading ? 'bg-blue-400 animate-pulse' :
                          ver.active ? 'bg-blue-400' :
                            ver.installed ? 'bg-green-400' :
                              'bg-gray-600'
                        }`} />

                      {/* Version label */}
                      <span className={`text-sm font-mono flex-1 ${ver.active ? 'text-blue-300 font-medium' :
                          ver.installed ? 'text-white' :
                            'text-gray-500'
                        }`}>
                        v{ver.version}
                      </span>

                      {/* Badges */}
                      <div className="flex items-center gap-2">
                        {ver.active && (
                          <span className="text-xs px-2 py-0.5 rounded-full bg-blue-500/20 text-blue-400 font-medium">
                            {t.bin_active}
                          </span>
                        )}
                        {ver.installed && !ver.active && (
                          <span className="text-xs text-gray-500">{t.bin_installed}</span>
                        )}
                      </div>

                      {/* Progress bar */}
                      {isDownloading && (
                        <div className="w-24">
                          <div className="h-1 bg-[#1e2535] rounded-full overflow-hidden">
                            <div
                              className="h-full bg-blue-500 rounded-full transition-all duration-200"
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <p className="text-xs text-blue-400 mt-0.5 text-right">
                            {progress === 0 ? t.bin_connecting : `${Math.round(progress)}%`}
                          </p>
                        </div>
                      )}

                      {/* Action buttons */}
                      <div className="flex items-center gap-1.5 flex-shrink-0">
                        {isDownloading ? (
                          <button
                            onClick={() => cancelDownload(svc.service, ver.version)}
                            className="px-2.5 py-1 rounded-lg text-xs bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                          >
                            {t.bin_cancel}
                          </button>
                        ) : ver.installed ? (
                          ver.active ? (
                            <span className="text-xs text-blue-400 px-2.5 py-1 inline-flex items-center gap-1">
                              <Check size={14} className="text-blue-400" /> {t.bin_active}
                            </span>
                          ) : (
                            <>
                              <button
                                onClick={() => setActiveVersion(svc.service, ver.version)}
                                className="px-2.5 py-1 rounded-lg text-xs bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors"
                              >
                                {t.bin_set_active}
                              </button>
                              {deleteConfirm === key ? (
                                <div className="flex items-center gap-1">
                                  <button
                                    onClick={async () => {
                                      try {
                                        await deleteBinary(svc.service, ver.version)
                                        setDeleteConfirm(null)
                                        setDeleteError('')
                                      } catch (e: any) {
                                        setDeleteError(e?.toString() ?? 'Delete failed')
                                        setDeleteConfirm(null)
                                      }
                                    }}
                                    className="px-2 py-1 rounded-lg text-xs bg-red-500 text-white hover:bg-red-600 transition-colors"
                                  >
                                    {t.yes}
                                  </button>
                                  <button
                                    onClick={() => setDeleteConfirm(null)}
                                    className="px-2 py-1 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                                  >
                                    {t.no}
                                  </button>
                                </div>
                              ) : (
                                <button
                                  onClick={() => setDeleteConfirm(key)}
                                  className="px-2.5 py-1 rounded-lg text-xs bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                                  title={t.delete}
                                >
                                  {t.delete}
                                </button>
                              )}
                            </>
                          )
                        ) : (
                          <button
                            onClick={() => downloadBinary(svc.service, ver.version)}
                            className="px-2.5 py-1 rounded-lg text-xs bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors"
                          >
                            {t.bin_download}
                          </button>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>

      <p className="text-xs text-gray-600">
        {t.bin_footer_1} <code className="text-gray-500">bin/{'<service>/<version>/'}</code>.
        {' '}{t.bin_footer_2} <strong className="text-gray-400">{t.bin_set_active}</strong> {t.bin_footer_3}
      </p>
    </div>
  )
}
