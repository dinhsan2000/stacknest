import { create } from 'zustand'
import { Locale, Translations } from './types'
import en from './en'
import vi from './vi'

const translations: Record<Locale, Translations> = { en, vi }

const STORAGE_KEY = 'stacknest_locale'

function getInitialLocale(): Locale {
    try {
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved === 'en' || saved === 'vi') return saved
    } catch { }
    // Auto-detect from browser language
    const lang = navigator.language.toLowerCase()
    if (lang.startsWith('vi')) return 'vi'
    return 'en'
}

interface I18nStore {
    locale: Locale
    t: Translations
    setLocale: (locale: Locale) => void
}

export const useI18n = create<I18nStore>((set) => {
    const initial = getInitialLocale()
    return {
        locale: initial,
        t: translations[initial],
        setLocale: (locale: Locale) => {
            localStorage.setItem(STORAGE_KEY, locale)
            set({ locale, t: translations[locale] })
        },
    }
})

/**
 * Template string helper: replaces {key} placeholders with values.
 * Usage: tt(t.dash_running_count, { running: 3, total: 5 })
 */
export function tt(template: string, vars: Record<string, string | number>): string {
    return template.replace(/\{(\w+)\}/g, (_, key) => String(vars[key] ?? `{${key}}`))
}

export type { Locale, Translations }
export const localeLabels: Record<Locale, string> = {
    en: 'English',
    vi: 'Tiếng Việt',
}
