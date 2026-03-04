import React, { useEffect, useState } from 'react'
import { Server, Database, Code2, Zap, Layers } from 'lucide-react'
import { useServiceStore } from '../store/serviceStore'
import { ServiceIcon } from '../components/ServiceIcon';

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

  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null) // "service@version"
  const [deleteError, setDeleteError] = useState('')

  useEffect(() => { fetchBinaryStatus() }, [])

  const missingCount = binaryStatus.filter(s => !s.versions.some(v => v.installed)).length

  return (
    <div className="flex flex-col gap-6 max-w-3xl">
      <div>
        <h2 className="text-2xl font-bold text-white">Service Binaries</h2>
        <p className="text-gray-400 text-sm mt-1">
          Tải và quản lý các phiên bản binary cho từng service.
          Binaries được lưu tại <code className="text-blue-400 text-xs">bin/{'{service}/{version}/'}</code>
        </p>
      </div>

      {missingCount > 0 && (
        <div className="flex items-center gap-3 bg-yellow-500/10 border border-yellow-500/20 rounded-xl px-4 py-3">
          <span className="text-yellow-400 text-lg mt-0.5">⚠</span>
          <p className="text-sm text-yellow-300">
            {missingCount} service{missingCount > 1 ? 's' : ''} chưa có binary được cài.
            Tải ít nhất một phiên bản trước khi khởi động.
          </p>
        </div>
      )}

      {/* Global download errors */}
      {Object.entries(downloadErrors).map(([key, errMsg]) => (
        <div key={key} className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3">
          <span className="text-red-400 text-lg">✕</span>
          <p className="text-sm text-red-300 flex-1">
            <strong>{key.replace('@', ' v')}</strong>: {errMsg}
          </p>
          <button
            onClick={() => dismissDownloadError(key)}
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            Dismiss
          </button>
        </div>
      ))}

      {deleteError && (
        <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3">
          <span className="text-red-400 text-lg">✕</span>
          <p className="text-sm text-red-300 flex-1">{deleteError}</p>
          <button
            onClick={() => setDeleteError('')}
            className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
          >
            Dismiss
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
                        {activeVersion ? `Active: v${activeVersion.version}` : 'Installed'}
                      </span>
                    ) : (
                      <span className="text-xs px-2 py-0.5 rounded-full bg-gray-500/15 text-gray-500">
                        Not installed
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
                            Active
                          </span>
                        )}
                        {ver.installed && !ver.active && (
                          <span className="text-xs text-gray-500">Installed</span>
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
                            {progress === 0 ? 'Connecting...' : `${Math.round(progress)}%`}
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
                            Cancel
                          </button>
                        ) : ver.installed ? (
                          ver.active ? (
                            <span className="text-xs text-blue-400 px-2.5 py-1">✓ Active</span>
                          ) : (
                            <>
                              <button
                                onClick={() => setActiveVersion(svc.service, ver.version)}
                                className="px-2.5 py-1 rounded-lg text-xs bg-green-500/10 text-green-400 hover:bg-green-500/20 transition-colors"
                              >
                                Set Active
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
                                    Yes
                                  </button>
                                  <button
                                    onClick={() => setDeleteConfirm(null)}
                                    className="px-2 py-1 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                                  >
                                    No
                                  </button>
                                </div>
                              ) : (
                                <button
                                  onClick={() => setDeleteConfirm(key)}
                                  className="px-2.5 py-1 rounded-lg text-xs bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                                  title="Delete this version"
                                >
                                  Delete
                                </button>
                              )}
                            </>
                          )
                        ) : (
                          <button
                            onClick={() => downloadBinary(svc.service, ver.version)}
                            className="px-2.5 py-1 rounded-lg text-xs bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors"
                          >
                            Download
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
        Mỗi phiên bản được lưu tại <code className="text-gray-500">bin/{'<service>/<version>/'}</code>.
        Sau khi tải xong, nhấn <strong className="text-gray-400">Set Active</strong> để chuyển sang phiên bản đó.
        Service đang chạy sẽ dùng phiên bản active trong lần restart tiếp theo.
      </p>
    </div>
  )
}
