import { useLocale } from '../../shared/i18n'

export function LanguageSwitcher() {
  const { locale, setLocale, t } = useLocale()

  return (
    <div className="btn-group" role="group" aria-label={t('languageAriaLabel')}>
      <button
        type="button"
        className={`btn btn-sm ${locale === 'id' ? 'btn-primary' : 'btn-outline-secondary'}`}
        onClick={() => setLocale('id')}
        aria-pressed={locale === 'id'}
      >
        ID
      </button>
      <button
        type="button"
        className={`btn btn-sm ${locale === 'en' ? 'btn-primary' : 'btn-outline-secondary'}`}
        onClick={() => setLocale('en')}
        aria-pressed={locale === 'en'}
      >
        EN
      </button>
    </div>
  )
}
