import { useLocale } from '../../shared/i18n'

const LOCALES = [
  { code: 'id', label: 'ID' },
  { code: 'en', label: 'EN' },
]

export function LanguageSwitcher() {
  const { locale, setLocale, t } = useLocale()

  return (
    <div className="ks-segmented" role="group" aria-label={t('languageAriaLabel')}>
      {LOCALES.map(({ code, label }) => (
        <button
          key={code}
          type="button"
          className={`ks-segment ${locale === code ? 'ks-segment-active' : ''}`}
          onClick={() => setLocale(code)}
          aria-pressed={locale === code}
        >
          {label}
        </button>
      ))}
    </div>
  )
}
