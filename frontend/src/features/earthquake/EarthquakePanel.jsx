import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

export function EarthquakePanel({ items = [], loading = false, error = null, onRetry }) {
  const { t } = useLocale()

  return (
    <div className="ks-card h-100" aria-labelledby="earthquakes-heading">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h2 id="earthquakes-heading" className="h5 mb-0 text-dark fw-bold">
          <i className="bi bi-activity me-2 text-danger"></i>
          {t('earthquakeTitle')}
        </h2>
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && items.length === 0 && (
        <AsyncEmpty message={t('emptyEarthquakes')} />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="d-flex flex-column gap-3">
          {items.map((eq, idx) => {
            const hasCoords = typeof eq.latitude === 'number' && typeof eq.longitude === 'number'

            return (
              <div
                key={eq.id || idx}
                className="p-3 border rounded"
                style={{ backgroundColor: '#fafbfc' }}
              >
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
                  <span className="fw-bold text-dark small">{eq.date_time || eq.dateTime || '-'}</span>
                  <div className="d-flex align-items-center gap-2">
                    {typeof eq.magnitude === 'number' && (
                      <span className="badge bg-danger">
                        M {eq.magnitude.toFixed(1)}
                      </span>
                    )}
                    {typeof eq.depth_km === 'number' && (
                      <span className="badge bg-secondary">
                        {eq.depth_km} km
                      </span>
                    )}
                  </div>
                </div>

                {eq.region && (
                  <div className="mb-2">
                    <span className="small text-muted">{t('region')}: </span>
                    <span className="small fw-semibold text-dark">{eq.region}</span>
                  </div>
                )}

                {hasCoords && (
                  <div className="small text-secondary mb-1">
                    <span>{t('coordinates')}: </span>
                    <span>{eq.latitude}, {eq.longitude}</span>
                  </div>
                )}

                {eq.potential && (
                  <div className="small text-secondary mb-1">
                    <span>{t('potential')}: </span>
                    <strong className="text-dark">{eq.potential}</strong>
                  </div>
                )}

                {eq.felt_areas && (
                  <div className="small text-secondary mt-1">
                    <span>{t('felt')}: </span>
                    <span>{eq.felt_areas}</span>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
