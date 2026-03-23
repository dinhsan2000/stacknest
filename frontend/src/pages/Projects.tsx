import { useEffect, useState } from 'react'
import { useServiceStore } from '../store/serviceStore'
import { useI18n } from '../i18n'
import { OpenFolder } from '../../wailsjs/go/main/App'
import {
  FolderKanban,
  Plus,
  Play,
  Trash2,
  FolderOpen,
  Globe,
  Lock,
  Server,
  Zap,
  Check,
  X,
  FileCode,
  Code2,
  Newspaper,
} from 'lucide-react'

export default function Projects() {
  const { t } = useI18n()
  const {
    projects, fetchProjects, quickCreateProject,
    applyProject, deactivateProject, deleteProject,
  } = useServiceStore()

  const [name, setName] = useState('')
  const [server, setServer] = useState<'apache' | 'nginx'>('apache')
  const [template, setTemplate] = useState<'blank' | 'laravel' | 'wordpress'>('blank')
  const [ssl, setSsl] = useState(false)
  const [creating, setCreating] = useState(false)
  const [applying, setApplying] = useState('')
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => { fetchProjects() }, [])

  const handleQuickCreate = async () => {
    if (!name.trim()) return
    setCreating(true)
    setError('')
    setSuccess('')
    try {
      await quickCreateProject(name.trim(), server, template, ssl)
      setSuccess(t.proj_created_ok)
      setName('')
      setTimeout(() => setSuccess(''), 3000)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to create project')
    } finally {
      setCreating(false)
    }
  }

  const handleApply = async (id: string) => {
    setApplying(id)
    setError('')
    setSuccess('')
    try {
      await applyProject(id)
      setSuccess(t.proj_applied_ok)
      setTimeout(() => setSuccess(''), 3000)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to apply project')
    } finally {
      setApplying('')
    }
  }

  const handleDelete = async (id: string) => {
    setConfirmDelete(null)
    try {
      await deleteProject(id)
    } catch (e: any) {
      setError(e?.toString() ?? 'Failed to delete project')
    }
  }

  const activeProject = projects.find(p => p.active)

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">{t.proj_title}</h2>
          {activeProject && (
            <p className="text-gray-400 text-sm mt-1">
              {t.proj_active}: <span className="text-green-400 font-semibold">{activeProject.name}</span>
            </p>
          )}
        </div>
      </div>

      {/* Alerts */}
      {error && (
        <div className="text-red-400 text-sm bg-red-500/10 rounded-lg px-4 py-3 flex items-center gap-2">
          <X size={14} className="flex-shrink-0" /> {error}
        </div>
      )}
      {success && (
        <div className="text-green-400 text-sm bg-green-500/10 rounded-lg px-4 py-3 flex items-center gap-2">
          <Check size={14} className="flex-shrink-0" /> {success}
        </div>
      )}

      {/* Quick Create */}
      <div className="bg-[#1e2535] border border-[#2a3347] rounded-xl p-5 flex flex-col gap-4">
        <div>
          <h3 className="text-white font-semibold">{t.proj_quick_create}</h3>
          <p className="text-gray-500 text-xs mt-1">{t.proj_quick_create_desc}</p>
        </div>

        {/* Template selector */}
        <div className="grid grid-cols-3 gap-2">
          {([
            { id: 'blank', label: t.proj_tpl_blank, desc: t.proj_tpl_blank_desc, icon: <FileCode size={20} /> },
            { id: 'laravel', label: t.proj_tpl_laravel, desc: t.proj_tpl_laravel_desc, icon: <Code2 size={20} /> },
            { id: 'wordpress', label: t.proj_tpl_wordpress, desc: t.proj_tpl_wordpress_desc, icon: <Newspaper size={20} /> },
          ] as const).map(tpl => (
            <button
              key={tpl.id}
              onClick={() => setTemplate(tpl.id)}
              className={`flex items-center gap-3 p-3 rounded-lg border text-left transition-colors ${
                template === tpl.id
                  ? 'border-blue-500/40 bg-blue-500/10 text-blue-400'
                  : 'border-[#2a3347] bg-[#0f1420] text-gray-400 hover:border-[#3a4357]'
              }`}
            >
              <span className="flex-shrink-0">{tpl.icon}</span>
              <div>
                <p className="text-sm font-medium">{tpl.label}</p>
                <p className="text-xs text-gray-500">{tpl.desc}</p>
              </div>
            </button>
          ))}
        </div>

        <div className="flex gap-3">
          <input
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleQuickCreate()}
            placeholder={t.proj_name_placeholder}
            className="flex-1 bg-[#0f1420] border border-[#2a3347] rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />

          {/* Server toggle */}
          <div className="flex items-center gap-1 bg-[#0f1420] border border-[#2a3347] rounded-lg p-0.5">
            {(['apache', 'nginx'] as const).map(srv => (
              <button
                key={srv}
                onClick={() => setServer(srv)}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors flex items-center gap-1 ${
                  server === srv
                    ? 'bg-blue-500/20 text-blue-400'
                    : 'text-gray-500 hover:text-gray-300'
                }`}
              >
                {srv === 'apache' ? <Server size={12} /> : <Zap size={12} />}
                {srv.charAt(0).toUpperCase() + srv.slice(1)}
              </button>
            ))}
          </div>

          {/* SSL toggle */}
          <label className="flex items-center gap-1.5 text-xs text-gray-400 cursor-pointer px-2">
            <input
              type="checkbox"
              checked={ssl}
              onChange={e => setSsl(e.target.checked)}
              className="accent-blue-500"
            />
            <Lock size={12} /> SSL
          </label>

          <button
            onClick={handleQuickCreate}
            disabled={creating || !name.trim()}
            className="px-4 py-2 rounded-lg bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-colors disabled:opacity-40 flex items-center gap-1.5"
          >
            <Plus size={14} />
            {creating ? t.proj_creating : t.proj_create}
          </button>
        </div>
      </div>

      {/* Project list */}
      {projects.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-gray-600 gap-3">
          <FolderKanban size={40} />
          <p>{t.proj_no_projects}</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {projects.map(proj => (
            <div
              key={proj.id}
              className={`bg-[#1e2535] border rounded-xl p-4 flex items-center gap-4 transition-all ${
                proj.active
                  ? 'border-green-500/30 ring-1 ring-green-500/20'
                  : 'border-[#2a3347] hover:border-[#3a4357]'
              }`}
            >
              {/* Icon */}
              <div className="flex-shrink-0">
                <FolderKanban size={20} className={proj.active ? 'text-green-400' : 'text-gray-500'} />
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-white font-semibold">{proj.name}</span>
                  {proj.active && (
                    <span className="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">
                      {t.proj_active}
                    </span>
                  )}
                  <span className="text-xs bg-[#0f1420] text-gray-500 px-2 py-0.5 rounded flex items-center gap-1">
                    {proj.server === 'apache' ? <Server size={10} /> : <Zap size={10} />}
                    {proj.server}
                  </span>
                  {proj.ssl && (
                    <span className="text-xs bg-[#0f1420] text-gray-500 px-2 py-0.5 rounded flex items-center gap-1">
                      <Lock size={10} /> SSL
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                  {proj.domain && (
                    <span className="flex items-center gap-1">
                      <Globe size={10} /> {proj.domain}
                    </span>
                  )}
                  <span className="truncate" title={proj.doc_root}>{proj.doc_root}</span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-2 flex-shrink-0">
                <button
                  onClick={() => OpenFolder(proj.doc_root)}
                  title={t.proj_open_folder}
                  className="px-2.5 py-1.5 rounded-lg text-xs bg-[#0f1420] text-gray-400 hover:text-white transition-colors"
                >
                  <FolderOpen size={14} />
                </button>

                {proj.active ? (
                  <button
                    onClick={() => deactivateProject()}
                    className="px-3 py-1.5 rounded-lg text-xs bg-gray-500/15 text-gray-400 hover:text-white transition-colors"
                  >
                    {t.proj_deactivate}
                  </button>
                ) : (
                  <button
                    onClick={() => handleApply(proj.id)}
                    disabled={!!applying}
                    className="px-3 py-1.5 rounded-lg text-xs font-medium bg-blue-500/15 text-blue-400 hover:bg-blue-500/25 transition-colors disabled:opacity-40 flex items-center gap-1"
                  >
                    <Play size={12} />
                    {applying === proj.id ? t.proj_applying : t.proj_apply}
                  </button>
                )}

                {confirmDelete === proj.id ? (
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleDelete(proj.id)}
                      className="px-2.5 py-1.5 rounded-lg text-xs bg-red-500 text-white hover:bg-red-600 transition-colors"
                    >
                      {t.yes}
                    </button>
                    <button
                      onClick={() => setConfirmDelete(null)}
                      className="px-2.5 py-1.5 rounded-lg text-xs bg-[#2a3347] text-gray-300 hover:bg-[#334060] transition-colors"
                    >
                      {t.no}
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setConfirmDelete(proj.id)}
                    title={t.proj_delete}
                    className="px-2.5 py-1.5 rounded-lg text-xs bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
