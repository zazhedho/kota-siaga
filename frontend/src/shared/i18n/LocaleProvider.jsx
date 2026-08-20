import { useEffect, useState } from 'react'
import { messages } from './messages'
import { readLocale, writeLocale } from '../storage'
import { LocaleContext } from './LocaleContext'

export function LocaleProvider({ children }) {
  const [locale, setLocaleState] = useState(readLocale())

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  const setLocale = (nextLocale) => {
    const next = nextLocale === 'en' ? 'en' : 'id'
    writeLocale(next)
    setLocaleState(next)
  }

  const t = (key, params) => {
    let text = messages[locale]?.[key] || messages.id[key] || key
    if (params && typeof params === 'object') {
      Object.entries(params).forEach(([paramKey, val]) => {
        text = text.replace(new RegExp(`\\{${paramKey}\\}`, 'g'), String(val))
      })
    }
    return text
  }

  return (
    <LocaleContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </LocaleContext.Provider>
  )
}
