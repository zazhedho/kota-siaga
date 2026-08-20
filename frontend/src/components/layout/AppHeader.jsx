import { useLocale } from '../../shared/i18n'
import { LanguageSwitcher } from '../common/LanguageSwitcher'

export function AppHeader() {
  const { t } = useLocale()

  return (
    <header className="ks-header py-3">
      <div className="container-fluid container-xl">
        <div className="d-flex flex-column flex-sm-row align-items-start align-items-sm-center justify-content-between gap-3">
          <div>
            <div className="d-flex align-items-center gap-2">
              <i className="bi bi-shield-check text-primary fs-3" aria-hidden="true"></i>
              <h1 className="h4 mb-0 ks-brand-title">{t('brandTitle')}</h1>
            </div>
            <p className="text-secondary small mb-0 mt-1">{t('brandSubtitle')}</p>
          </div>
          <div className="align-self-end align-self-sm-center">
            <LanguageSwitcher />
          </div>
        </div>
      </div>
    </header>
  )
}
