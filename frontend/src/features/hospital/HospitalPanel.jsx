import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

export function HospitalPanel({
  hospitals = [],
  total = 0,
  loading = false,
  loadingMore = false,
  error = null,
  hasNextPage = false,
  onLoadMore,
  onRetry,
}) {
  const { t } = useLocale()

  return (
    <div className="ks-card h-100" aria-labelledby="hospitals-heading">
      <div className="mb-3">
        <div className="d-flex align-items-center justify-content-between">
          <h2 id="hospitals-heading" className="h5 mb-0 text-dark fw-bold">
            <i className="bi bi-hospital me-2 text-primary"></i>
            {t('hospitalTitle')}
          </h2>
          {total > 0 && (
            <span className="badge bg-light text-dark border">
              {t('showingHospitalsCount', { count: hospitals.length, total })}
            </span>
          )}
        </div>
        <p className="small text-secondary mb-0 mt-1">{t('hospitalSubtitle')}</p>
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && hospitals.length === 0 && (
        <AsyncEmpty message={t('emptyHospitals')} />
      )}

      {!loading && !error && hospitals.length > 0 && (
        <div>
          <div className="d-flex flex-column gap-2 mb-3">
            {hospitals.map((h, idx) => (
              <div key={h.id || idx} className="ks-hospital-card">
                <div className="d-flex flex-wrap align-items-start justify-content-between gap-2 mb-1">
                  <h3 className="h6 mb-0 fw-bold text-dark">{h.name}</h3>
                  <div className="d-flex gap-1 flex-wrap">
                    {h.type && <span className="badge bg-info-subtle text-info-emphasis border">{h.type}</span>}
                    {h.class && <span className="badge bg-secondary-subtle text-secondary-emphasis border">{t('hospitalClass', { value: h.class })}</span>}
                  </div>
                </div>

                {h.address && (
                  <p className="small text-secondary mb-1">
                    <i className="bi bi-geo-alt me-1 text-muted" aria-hidden="true"></i>
                    {h.address}
                    {h.postal_code ? ` ${h.postal_code}` : ''}
                  </p>
                )}

                <div className="d-flex flex-wrap align-items-center gap-3 small text-muted">
                  {h.phone && (
                    <span>
                      <i className="bi bi-telephone me-1" aria-hidden="true"></i>
                      <a href={`tel:${h.phone}`} className="text-decoration-none text-dark">
                        {h.phone}
                      </a>
                    </span>
                  )}
                  {typeof h.beds_total === 'number' && h.beds_total > 0 && (
                    <span>
                      <i className="bi bi-door-closed me-1" aria-hidden="true"></i>
                      {t('hospitalBeds', { count: h.beds_total })}
                    </span>
                  )}
                  {typeof h.icu_beds === 'number' && h.icu_beds > 0 && (
                    <span>
                      <i className="bi bi-heart-pulse me-1 text-danger" aria-hidden="true"></i>
                      {t('hospitalICUBeds', { count: h.icu_beds })}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>

          {hasNextPage && (
            <div className="text-center pt-2">
              <button
                type="button"
                className="btn btn-outline-primary btn-sm px-4"
                onClick={onLoadMore}
                disabled={loadingMore}
                aria-busy={loadingMore}
              >
                {loadingMore ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true" />
                    {t('loading')}
                  </>
                ) : (
                  <>
                    <i className="bi bi-arrow-down-circle me-1" aria-hidden="true"></i>
                    {t('loadMore')}
                  </>
                )}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
