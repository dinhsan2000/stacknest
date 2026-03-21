import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import {
  SelectFolder,
  GetSSLCerts,
  IsSSLCAInstalled,
  TrustSSLCA,
  GenerateSSLCert,
  GetCACertPath,
  OpenFolder,
} from '../../wailsjs/go/main/App'

interface CertInfo {
  domain:     string
  cert_path:  string
  key_path:   string
  expires_at: string
}

export default function VirtualHosts() {
  const { vhosts, fetchVHosts, addVHost, removeVHost } = useServiceStore()
  const [form, setForm]         = useState({ name: '', domain: '', root: '', server: 'apache', ssl: false })
  const [error, setError]       = useState('')
  const [confirmDomain, setConfirmDomain] = useState<string | null>(null)

  // SSL state
  const [caInstalled, setCAInstalled] = useState(false)
  const [certs, setCerts]             = useState<CertInfo[]>([])
  const [trustingCA, setTrustingCA]   = useState(false)
  const [genDomain, setGenDomain]     = useState('')   // domain currently generating cert
  const [sslError, setSSLError]       = useState('')
  const [sslSuccess, setSSLSuccess]   = useState('')

  const refreshSSL = async () => {
    const [installed, certList] = await Promise.all([
      IsSSLCAInstalled(),
      GetSSLCerts(),
    ])
    setCAInstalled(installed)
    setCerts((certList || []) as CertInfo[])
  }

  useEffect(() => {
    fetchVHosts()
    refreshSSL()
  }, [])

  const handleSelectFolder = async () => {
    const path = await SelectFolder()
    if (path) setForm(f => ({ ...f, root: path }))
  }

  const handleAdd = async () => {
    if (!form.domain || !form.root) {
      setError('Domain and root path are required')
      return
    }
    try {
      await addVHost(form.name || form.domain, form.domain, form.root, form.server, form.ssl)
      setForm({ name: '', domain: '', root: '', server: 'apache', ssl: false })
      setError('')
    } catch (e: any) {
      setError(e.toString())
    }
  }

  const handleTrustCA = async () => {
    setTrustingCA(true)
    setSSLError('')
    setSSLSuccess('')
    try {
      await TrustSSLCA()
      setSSLSuccess('CA certificate trusted successfully!')
      await refreshSSL()
    } catch (e: any) {
      setSSLError(e?.toString() ?? 'Failed to trust CA')
    } finally {
      setTrustingCA(false)
    }
  }

  const handleExportCA = async () => {
    const path = await GetCACertPath()
    if (path) OpenFolder(path.replace(/[^/\\]+$/, '')) // open containing folder
  }

  const handleGenerateCert = async (domain: string) => {
    setGenDomain(domain)
    setSSLError('')
    setSSLSuccess('')
    try {
      await GenerateSSLCert(domain)
      setSSLSuccess(`Certificate generated for ${domain}`)
      await refreshSSL()
    } catch (e: any) {
      setSSLError(e?.toString() ?? 'Failed to generate certificate')
    } finally {
      setGenDomain('')
    }
  }

  const certFor = (domain: string) => certs.find(c => c.domain === domain)

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      <h2 className="text-2xl font-bold text-white">Virtual Hosts</h2>

      {/* CA Trust Banner */}
      <div className={`flex items-center gap-4 px-5 py-3 rounded-xl border
        ${caInstalled
          ? 'bg-green-500/10 border-green-500/30'
          : 'bg-yellow-500/10 border-yellow-500/30'
        }`}
      >
        <span className="text-xl">{caInstalled ? '🔒' : '🔓'}</span>
        <div className="flex-1">
          <p className="text-sm font-medium text-white">Local SSL CA</p>
          <p className={`text-xs mt-0.5 ${caInstalled ? 'text-green-400' : 'text-yellow-400'}`}>
            {caInstalled
              ? 'CA is trusted by this machine — HTTPS will work without browser warnings'
              : 'CA is not trusted — browsers will show SSL warnings until you trust it'}
          </p>
        </div>
        <div className="flex gap-2">
          {caInstalled ? (
            <button
              onClick={handleExportCA}
              className="px-3 py-1.5 text-xs rounded-lg bg-[#1e2535] text-gray-400 hover:text-white transition-colors"
            >
              Export CA Cert
            </button>
          ) : (
            <button
              onClick={handleTrustCA}
              disabled={trustingCA}
              className="px-4 py-1.5 text-xs rounded-lg bg-yellow-500 text-black font-semibold hover:bg-yellow-400 transition-colors disabled:opacity-50"
            >
              {trustingCA ? 'Trusting...' : 'Trust CA'}
            </button>
          )}
        </div>
      </div>

      {/* SSL alerts */}
      {sslError   && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-4 py-2">{sslError}</p>}
      {sslSuccess && <p className="text-green-400 text-sm bg-green-500/10 rounded-lg px-4 py-2">✓ {sslSuccess}</p>}

      {/* Add form */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5 flex flex-col gap-4">
        <h3 className="text-white font-semibold">Add Virtual Host</h3>

        <div className="grid grid-cols-2 gap-3">
          <input
            placeholder="Name (optional)"
            value={form.name}
            onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
            className="bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <input
            placeholder="Domain (e.g. myapp.test)"
            value={form.domain}
            onChange={e => setForm(f => ({ ...f, domain: e.target.value }))}
            className="bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        <div className="flex gap-3">
          <input
            placeholder="Document root path"
            value={form.root}
            onChange={e => setForm(f => ({ ...f, root: e.target.value }))}
            className="flex-1 bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
          <button
            onClick={handleSelectFolder}
            className="px-3 py-2 rounded-lg bg-[#0f1420] border border-[#2a3347] text-gray-400 hover:text-white text-sm transition-colors"
          >
            Browse
          </button>
        </div>

        <div className="flex items-center justify-between gap-4">
          {/* Server toggle */}
          <div className="flex items-center gap-1 bg-[#0f1420] border border-[#2a3347] rounded-lg p-0.5">
            {(['apache', 'nginx'] as const).map(srv => (
              <button
                key={srv}
                onClick={() => setForm(f => ({ ...f, server: srv }))}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors capitalize ${
                  form.server === srv
                    ? 'bg-blue-500 text-white'
                    : 'text-gray-400 hover:text-white'
                }`}
              >
                {srv}
              </button>
            ))}
          </div>

          <label className="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
            <input
              type="checkbox"
              checked={form.ssl}
              onChange={e => setForm(f => ({ ...f, ssl: e.target.checked }))}
              className="accent-blue-500"
            />
            Enable SSL (HTTPS)
          </label>

          <button
            onClick={handleAdd}
            className="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 text-sm font-medium transition-colors"
          >
            Add Host
          </button>
        </div>

        {error && <p className="text-red-400 text-xs">{error}</p>}
      </div>

      {/* Hosts list */}
      <div className="flex flex-col gap-2">
        {vhosts.length === 0 && (
          <p className="text-gray-500 text-center py-8">No virtual hosts configured yet.</p>
        )}
        {vhosts.map(vh => {
          const cert = certFor(vh.domain)
          return (
            <div
              key={vh.domain}
              className="bg-[#1e2535] border border-[#2a3347] rounded-lg px-5 py-3 flex items-center justify-between gap-4"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-white font-medium">{vh.domain}</span>
                  <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${
                    vh.server === 'nginx'
                      ? 'bg-green-500/20 text-green-400'
                      : 'bg-orange-500/20 text-orange-400'
                  }`}>
                    {vh.server ?? 'apache'}
                  </span>
                  {vh.ssl && (
                    <span className="text-xs bg-blue-500/20 text-blue-400 px-1.5 py-0.5 rounded">SSL</span>
                  )}
                </div>
                <p className="text-xs text-gray-500 mt-0.5">{vh.root}</p>

                {/* SSL cert info */}
                {vh.ssl && (
                  <div className="flex items-center gap-3 mt-1.5">
                    {cert ? (
                      <span className="text-xs text-gray-400">
                        🔐 Cert expires: <span className="text-gray-300">{cert.expires_at}</span>
                      </span>
                    ) : (
                      <span className="text-xs text-yellow-500">⚠ No certificate yet</span>
                    )}
                    <button
                      onClick={() => handleGenerateCert(vh.domain)}
                      disabled={genDomain === vh.domain}
                      className="text-xs text-blue-400 hover:text-blue-300 transition-colors disabled:opacity-50"
                    >
                      {genDomain === vh.domain
                        ? 'Generating...'
                        : cert ? '↻ Regenerate Cert' : '+ Generate Cert'}
                    </button>
                  </div>
                )}
              </div>

              <div className="flex gap-2 flex-shrink-0 items-center">
                <a
                  href={`http${vh.ssl ? 's' : ''}://${vh.domain}`}
                  className="px-3 py-1.5 rounded-lg text-xs bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors"
                >
                  Open
                </a>
                {confirmDomain === vh.domain ? (
                  <>
                    <span className="text-xs text-red-400">Remove this host?</span>
                    <button
                      onClick={() => { removeVHost(vh.domain); setConfirmDomain(null) }}
                      className="px-3 py-1.5 rounded-lg text-xs bg-red-500 text-white hover:bg-red-600 transition-colors"
                    >
                      Yes, remove
                    </button>
                    <button
                      onClick={() => setConfirmDomain(null)}
                      className="px-3 py-1.5 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                    >
                      Cancel
                    </button>
                  </>
                ) : (
                  <button
                    onClick={() => setConfirmDomain(vh.domain)}
                    className="px-3 py-1.5 rounded-lg text-xs bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
                  >
                    Remove
                  </button>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
