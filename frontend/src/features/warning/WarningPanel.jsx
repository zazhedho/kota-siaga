import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

function getSeverityDetails(severity, t) {
  const norm = (severity || '').toLowerCase()
  if (norm.includes('extreme') || norm.includes('ekstrem') || norm.includes('kritis')) {
    return {
      cardClass: 'ks-warning-card-extreme',
      badgeClass: 'ks-badge-danger',
      icon: 'bi-exclamation-octagon-fill',
      label: t('severityExtreme'),
    }
  }
  if (norm.includes('severe') || norm.includes('awas') || norm.includes('berat')) {
    return {
      cardClass: 'ks-warning-card-severe',
      badgeClass: 'ks-badge-danger',
      icon: 'bi-exclamation-diamond-fill',
      label: t('severitySevere'),
    }
  }
  if (norm.includes('moderate') || norm.includes('siaga') || norm.includes('sedang')) {
    return {
      cardClass: 'ks-warning-card-moderate',
      badgeClass: 'ks-badge-warning',
      icon: 'bi-exclamation-triangle-fill',
      label: t('severityModerate'),
    }
  }
  if (norm.includes('minor') || norm.includes('waspada')) {
    return {
      cardClass: 'ks-warning-card-minor',
      badgeClass: 'ks-badge-warning',
      icon: 'bi-exclamation-circle-fill',
      label: t('severityMinor'),
    }
  }
  return {
    cardClass: '',
    badgeClass: 'ks-badge-neutral',
    icon: 'bi-info-circle-fill',
    label: severity || t('severityUnknown'),
  }
}

export function WarningPanel({ warnings = [], loading = false, error = null, onRetry }) {
  const { t } = useLocale()

  return (
    <div className="ks-card" aria-labelledby="warnings-heading">
      <div className="d-flex align-items-center justify-content-between ks-card-header">
        <div className="d-flex align-items-center gap-2">
          <div className="p-2 rounded-circle bg-warning-subtle text-warning-emphasis d-inline-flex">
            <i className="bi bi-shield-exclamation fs-5" aria-hidden="true"></i>
          </div>
          <div>
            <h2 id="warnings-heading" className="h6 mb-0 text-dark fw-bold">
              {t('warningTitle')}
            </h2>
            <span className="text-secondary" style={{ fontSize: '0.75rem' }}>
              {warnings.length > 0
                ? t('activeWarningsCount', { count: warnings.length })
                : t('normalCondition')}
            </span>
          </div>
        </div>
        {warnings.length > 0 ? (
          <span className="badge bg-warning text-dark px-2 py-1 rounded-pill fw-bold">
            {t('activeWarningCount', { count: warnings.length })}
          </span>
        ) : (
          <span className="badge bg-success-subtle text-success-emphasis border border-success-subtle px-2 py-1 rounded-pill small">
            <i className="bi bi-check-circle-fill me-1" aria-hidden="true"></i>
            {t('normalCondition')}
          </span>
        )}
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && warnings.length === 0 && (
        <AsyncEmpty
          variant="safe"
          icon="bi-shield-check"
          title={t('emptyWarningsTitle')}
          message={t('emptyWarnings')}
          badge={
            <span className="badge bg-success text-white px-3 py-1 rounded-pill small fw-semibold shadow-xs">
              <i className="bi bi-check2 me-1" aria-hidden="true"></i>
              {t('safeStatusBadge')}
            </span>
          }
        />
      )}

      {!loading && !error && warnings.length > 0 && (
        <div className={`d-flex flex-column gap-3 ${warnings.length > 3 ? 'ks-scrollable-list' : ''}`}>
          {warnings.map((w, idx) => {
            const details = getSeverityDetails(w.severity, t)
            return (
              <div
                key={w.id || idx}
                className={`ks-warning-card ${details.cardClass}`}
              >
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
                  <h3 className="h6 mb-0 fw-bold text-dark">{w.event || w.headline || t('warningFallbackTitle')}</h3>
                  <span className={`badge ${details.badgeClass} d-inline-flex align-items-center gap-2 px-2 py-1 rounded-pill`}>
                    <i className={`bi ${details.icon}`} aria-hidden="true"></i>
                    <span>{details.label}</span>
                  </span>
                </div>

                {w.headline && w.headline !== w.event && (
                  <p className="small text-dark mb-2 fw-semibold">{w.headline}</p>
                )}

                {w.description && (
                  <p className="small text-secondary mb-2 lh-base">{w.description}</p>
                )}

                {w.instruction && (
                  <div className="small p-2 bg-white border border-secondary-subtle rounded mb-2 text-dark">
                    <strong className="text-dark d-block mb-0.5">
                      <i className="bi bi-info-circle me-1 text-primary"></i>
                      {t('instructionLabel')}:
                    </strong>
                    <span>{w.instruction}</span>
                  </div>
                )}

                <div className="d-flex flex-wrap align-items-center gap-3 small text-muted pt-2 border-top">
                  {w.effective && (
                    <span className="d-inline-flex align-items-center gap-1">
                      <i className="bi bi-clock text-secondary"></i>
                      <span>{t('effectiveTime')}:</span>
                      <strong className="text-dark">{w.effective}</strong>
                    </span>
                  )}
                  {w.expires && (
                    <span className="d-inline-flex align-items-center gap-1">
                      <i className="bi bi-hourglass-split text-secondary"></i>
                      <span>{t('expiryTime')}:</span>
                      <strong className="text-dark">{w.expires}</strong>
                    </span>
                  )}
                  {w.source && (
                    <span className="ms-auto small text-muted">
                      {t('sourceLabel')}: {w.source}
                    </span>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
