import { useLocale } from '../../shared/i18n'

export function AsyncLoading({ message }) {
  const { t } = useLocale()
  return (
    <div className="py-4 text-center text-secondary" role="status" aria-live="polite">
      <div className="spinner-border spinner-border-sm text-primary me-2" aria-hidden="true" />
      <span>{message || t('loading')}</span>
    </div>
  )
}

export function AsyncError({ message, onRetry }) {
  const { t } = useLocale()
  return (
    <div className="alert alert-danger my-3 d-flex flex-column flex-sm-row justify-content-between align-items-sm-center gap-2" role="alert">
      <div>
        <i className="bi bi-exclamation-triangle-fill me-2" aria-hidden="true"></i>
        <span>{message || t('genericError')}</span>
      </div>
      {onRetry && (
        <button
          type="button"
          className="btn btn-outline-danger btn-sm text-nowrap align-self-start align-self-sm-center"
          onClick={onRetry}
        >
          <i className="bi bi-arrow-clockwise me-1" aria-hidden="true"></i>
          {t('retry')}
        </button>
      )}
    </div>
  )
}

export function AsyncEmpty({ message }) {
  return (
    <div className="py-4 text-center text-muted">
      <i className="bi bi-info-circle fs-4 d-block mb-2 text-secondary" aria-hidden="true"></i>
      <p className="mb-0 small">{message}</p>
    </div>
  )
}
