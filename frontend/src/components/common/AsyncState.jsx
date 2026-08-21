import { useLocale } from '../../shared/i18n'

export function AsyncLoading({ message }) {
  const { t } = useLocale()
  return (
    <div className="py-5 text-center text-secondary" role="status" aria-live="polite">
      <div
        className="spinner-border spinner-border-sm text-primary me-2"
        style={{ width: '1.25rem', height: '1.25rem', borderWidth: '0.15em' }}
        aria-hidden="true"
      />
      <span className="fw-medium small">{message || t('loading')}</span>
    </div>
  )
}

export function AsyncError({ message, onRetry }) {
  const { t } = useLocale()
  return (
    <div
      className="alert alert-danger my-3 d-flex flex-column flex-sm-row justify-content-between align-items-sm-center gap-3 border-danger-subtle bg-danger-subtle text-danger-emphasis shadow-sm"
      style={{ borderRadius: 'var(--ks-radius-sm)' }}
      role="alert"
    >
      <div className="d-flex align-items-start gap-2">
        <i className="bi bi-exclamation-octagon-fill fs-5 text-danger mt-n1" aria-hidden="true"></i>
        <div>
          <strong className="d-block small">{t('genericError')}</strong>
          <span className="small">{message || t('genericError')}</span>
        </div>
      </div>
      {onRetry && (
        <button
          type="button"
          className="btn btn-danger btn-sm text-nowrap px-3 align-self-start align-self-sm-center shadow-sm"
          onClick={onRetry}
        >
          <i className="bi bi-arrow-clockwise me-1" aria-hidden="true"></i>
          {t('retry')}
        </button>
      )}
    </div>
  )
}

export function AsyncEmpty({
  icon = 'bi-info-circle',
  title,
  message,
  variant = 'default',
  badge,
  children,
}) {
  const isSafe = variant === 'safe'
  const isInfo = variant === 'info'

  let iconClass = 'ks-empty-icon-neutral'
  if (isSafe) iconClass = 'ks-empty-icon-success'
  else if (isInfo) iconClass = 'ks-empty-icon-info'

  return (
    <div className={`ks-empty-container ${isSafe ? 'ks-empty-container-safe' : ''}`}>
      <div className={`ks-empty-icon-wrap ${iconClass}`}>
        <i className={`bi ${icon}`} aria-hidden="true"></i>
      </div>
      {badge && <div className="mb-2">{badge}</div>}
      {title && <h3 className="h6 fw-bold text-dark mb-1">{title}</h3>}
      <p className="small text-secondary mb-0 mx-auto" style={{ maxWidth: '420px', lineHeight: 1.5 }}>
        {message}
      </p>
      {children && <div className="mt-3">{children}</div>}
    </div>
  )
}
