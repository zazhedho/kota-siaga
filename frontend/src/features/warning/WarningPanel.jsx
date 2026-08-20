import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

function getSeverityBadge(severity, t) {
  const norm = (severity || '').toLowerCase()
  if (norm.includes('extreme') || norm.includes('ekstrem') || norm.includes('kritis')) {
    return {
      className: 'ks-badge-danger',
      icon: 'bi-exclamation-octagon-fill',
      label: t('severityExtreme'),
    }
  }
  if (norm.includes('severe') || norm.includes('awas') || norm.includes('berat')) {
    return {
      className: 'ks-badge-danger',
      icon: 'bi-exclamation-diamond-fill',
      label: t('severitySevere'),
    }
  }
  if (norm.includes('moderate') || norm.includes('siaga') || norm.includes('sedang')) {
    return {
      className: 'ks-badge-warning',
      icon: 'bi-exclamation-triangle-fill',
      label: t('severityModerate'),
    }
  }
  if (norm.includes('minor') || norm.includes('waspada')) {
    return {
      className: 'ks-badge-warning',
      icon: 'bi-exclamation-circle-fill',
      label: t('severityMinor'),
    }
  }
  return {
    className: 'ks-badge-neutral',
    icon: 'bi-info-circle-fill',
    label: severity || t('severityUnknown'),
  }
}

export function WarningPanel({ warnings = [], loading = false, error = null, onRetry }) {
  const { t } = useLocale()

  return (
    <div className="ks-card h-100" aria-labelledby="warnings-heading">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h2 id="warnings-heading" className="h5 mb-0 text-dark fw-bold">
          <i className="bi bi-shield-exclamation me-2 text-warning"></i>
          {t('warningTitle')}
        </h2>
        {warnings.length > 0 && (
          <span className="badge bg-warning text-dark">{warnings.length}</span>
        )}
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && warnings.length === 0 && (
        <AsyncEmpty message={t('emptyWarnings')} />
      )}

      {!loading && !error && warnings.length > 0 && (
        <div className="d-flex flex-column gap-3">
          {warnings.map((w, idx) => {
            const badge = getSeverityBadge(w.severity, t)
            return (
              <div
                key={w.id || idx}
                className="p-3 border rounded"
                style={{ backgroundColor: '#fafbfc' }}
              >
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
                  <h3 className="h6 mb-0 fw-bold text-dark">{w.event || w.headline || t('warningFallbackTitle')}</h3>
                  <span className={`badge ${badge.className} d-inline-flex align-items-center gap-1`}>
                    <i className={`bi ${badge.icon}`} aria-hidden="true"></i>
                    <span>{badge.label}</span>
                  </span>
                </div>

                {w.headline && w.headline !== w.event && (
                  <p className="small text-dark mb-2 fw-semibold">{w.headline}</p>
                )}

                {w.description && (
                  <p className="small text-secondary mb-2">{w.description}</p>
                )}

                {w.instruction && (
                  <div className="small p-2 bg-white border rounded mb-2 text-dark">
                    <strong>{t('instructionLabel')}:</strong> {w.instruction}
                  </div>
                )}

                <div className="d-flex flex-wrap align-items-center gap-3 small text-muted pt-1 border-top">
                  {w.effective && (
                    <span>
                      {t('effectiveTime')}: <strong className="text-dark">{w.effective}</strong>
                    </span>
                  )}
                  {w.expires && (
                    <span>
                      {t('expiryTime')}: <strong className="text-dark">{w.expires}</strong>
                    </span>
                  )}
                  {w.source && (
                    <span>
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
