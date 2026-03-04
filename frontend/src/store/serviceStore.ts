import { create } from 'zustand'
import { ServiceInfo, VirtualHost, AppConfig, ServiceVersionStatus } from '../types'
import {
  GetServices,
  StartService,
  StopService,
  RestartService,
  StartAll,
  StopAll,
  GetVirtualHosts,
  AddVirtualHost,
  RemoveVirtualHost,
  GetConfig,
  GetBinaryStatus,
  StartBinaryDownload,
  SetActiveVersion,
  SetServiceEnabled,
  CancelBinaryDownload,
  DeleteBinary,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface ServiceStore {
  services: ServiceInfo[]
  vhosts: VirtualHost[]
  config: AppConfig | null
  loading: Record<string, boolean>
  binaryStatus: ServiceVersionStatus[]
  // key: "service@version", value: download progress 0-100
  downloadProgress: Record<string, number>
  // key: "service@version", value: error message
  downloadErrors: Record<string, string>

  fetchServices: () => Promise<void>
  startService: (name: string) => Promise<void>
  stopService: (name: string) => Promise<void>
  restartService: (name: string) => Promise<void>
  startAll: () => void
  stopAll: () => void

  fetchVHosts: () => Promise<void>
  addVHost: (name: string, domain: string, root: string, server: string, ssl: boolean) => Promise<void>
  removeVHost: (domain: string) => Promise<void>

  fetchConfig: () => Promise<void>
  fetchBinaryStatus: () => Promise<void>
  downloadBinary: (service: string, version: string) => void
  cancelDownload: (service: string, version: string) => void
  deleteBinary: (service: string, version: string) => Promise<void>
  setActiveVersion: (service: string, version: string) => Promise<void>
  setServiceEnabled: (name: string, enabled: boolean) => Promise<void>
  dismissDownloadError: (key: string) => void
  initEventListeners: () => void
}

export const useServiceStore = create<ServiceStore>((set, get) => ({
  services: [],
  vhosts: [],
  config: null,
  loading: {},
  binaryStatus: [],
  downloadProgress: {},
  downloadErrors: {},

  fetchServices: async () => {
    const services = await GetServices()
    set({ services: (services || []) as ServiceInfo[] })
  },

  startService: async (name) => {
    set(s => ({ loading: { ...s.loading, [name]: true } }))
    try {
      await StartService(name)
      await get().fetchServices()
    } finally {
      set(s => ({ loading: { ...s.loading, [name]: false } }))
    }
  },

  stopService: async (name) => {
    set(s => ({ loading: { ...s.loading, [name]: true } }))
    try {
      await StopService(name)
      await get().fetchServices()
    } finally {
      set(s => ({ loading: { ...s.loading, [name]: false } }))
    }
  },

  restartService: async (name) => {
    set(s => ({ loading: { ...s.loading, [name]: true } }))
    try {
      await RestartService(name)
      await get().fetchServices()
    } finally {
      set(s => ({ loading: { ...s.loading, [name]: false } }))
    }
  },

  startAll: () => {
    StartAll()
    setTimeout(() => get().fetchServices(), 1000)
  },

  stopAll: () => {
    StopAll()
    setTimeout(() => get().fetchServices(), 500)
  },

  fetchVHosts: async () => {
    const vhosts = await GetVirtualHosts()
    set({ vhosts: vhosts || [] })
  },

  addVHost: async (name, domain, root, server, ssl) => {
    await AddVirtualHost(name, domain, root, server, ssl)
    await get().fetchVHosts()
  },

  removeVHost: async (domain) => {
    await RemoveVirtualHost(domain)
    await get().fetchVHosts()
  },

  fetchConfig: async () => {
    const config = await GetConfig()
    set({ config: config as unknown as AppConfig })
  },

  fetchBinaryStatus: async () => {
    const status = await GetBinaryStatus()
    set({ binaryStatus: (status || []) as ServiceVersionStatus[] })
  },

  downloadBinary: (service, version) => {
    const key = `${service}@${version}`
    // Clear previous error for this key
    set(s => ({
      downloadProgress: { ...s.downloadProgress, [key]: 0 },
      downloadErrors: { ...s.downloadErrors, [key]: undefined } as any,
    }))
    StartBinaryDownload(service, version).catch(() => {
      set(s => {
        const { [key]: _, ...rest } = s.downloadProgress
        return { downloadProgress: rest }
      })
    })
  },

  cancelDownload: (service, version) => {
    const key = `${service}@${version}`
    CancelBinaryDownload(service, version)
    set(s => {
      const { [key]: _, ...rest } = s.downloadProgress
      return { downloadProgress: rest }
    })
  },

  deleteBinary: async (service, version) => {
    await DeleteBinary(service, version)
    await get().fetchBinaryStatus()
  },

  setActiveVersion: async (service, version) => {
    await SetActiveVersion(service, version)
    await get().fetchBinaryStatus()
  },

  setServiceEnabled: async (name, enabled) => {
    await SetServiceEnabled(name, enabled)
    await get().fetchServices()
  },

  dismissDownloadError: (key: string) => {
    set(s => {
      const { [key]: _, ...rest } = s.downloadErrors
      return { downloadErrors: rest }
    })
  },

  initEventListeners: () => {
    // Prevent duplicate registration (React StrictMode calls useEffect twice)
    if ((window as any).__stacknest_events_init) return
      ; (window as any).__stacknest_events_init = true

    EventsOn('services:updated', (services: unknown[]) => {
      set({ services: (services || []) as ServiceInfo[] })
    })

    EventsOn('binary:progress', (data: { service: string; version: string; pct: number }) => {
      const key = `${data.service}@${data.version}`
      set(s => ({ downloadProgress: { ...s.downloadProgress, [key]: data.pct } }))
    })

    EventsOn('binary:done', (data: { service: string; version: string; error: string }) => {
      const key = `${data.service}@${data.version}`
      set(s => {
        const { [key]: _, ...rest } = s.downloadProgress
        const errors = data.error
          ? { ...s.downloadErrors, [key]: data.error }
          : s.downloadErrors
        return { downloadProgress: rest, downloadErrors: errors }
      })
      get().fetchBinaryStatus()
    })
  },
}))
