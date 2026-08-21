import { useLocale } from '../../shared/i18n'
import { LanguageSwitcher } from '../common/LanguageSwitcher'

export function AppHeader() {
  const { t } = useLocale()

  return (
    <header className="ks-header sticky-top">
      <div className="container-fluid container-xl">
        <div className="ks-header-bar">
          <div className="ks-header-brand">
            <span className="ks-brand-badge" aria-hidden="true">
              <i className="bi bi-shield-check"></i>
            </span>
            <div className="ks-brand-text">
              <h1 className="ks-brand-title">{t('brandTitle')}</h1>
              <p className="ks-brand-subtitle">
                <span className="ks-live-dot" aria-hidden="true"></span>
                <span className="ks-live-label">{t('livePortalBadge')}</span>
                <span className="ks-brand-tagline">
                  <span aria-hidden="true"> &middot; </span>
                  {t('brandSubtitle')}
                </span>
              </p>
            </div>
          </div>

          <div className="ks-header-actions">
            <div
              className="ks-segmented ks-segmented-danger"
              role="group"
              aria-label={t('emergencyGroupAriaLabel')}
            >
              <a href="tel:112" className="ks-segment" aria-label={t('emergencyCall112')}>
                <i className="bi bi-shield-shaded" aria-hidden="true"></i>
                <span>112</span>
              </a>
              <a href="tel:119" className="ks-segment" aria-label={t('emergencyCall119')}>
                <i className="bi bi-heart-pulse-fill" aria-hidden="true"></i>
                <span>119</span>
              </a>
            </div>
            <LanguageSwitcher />
          </div>
        </div>
      </div>
    </header>
  )
}
