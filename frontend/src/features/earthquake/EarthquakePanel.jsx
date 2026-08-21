import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'
import { formatEarthquakeDateTime } from './earthquakeTime'

export function EarthquakePanel({ items = [], loading = false, error = null, onRetry }) {
  const { locale, t } = useLocale()

  return (
    <div className="ks-card" aria-labelledby="earthquakes-heading">
      <div className="d-flex align-items-center justify-content-between ks-card-header">
        <div className="d-flex align-items-center gap-2">
          <div className="p-2 rounded-circle bg-danger-subtle text-danger d-inline-flex">
            <i className="bi bi-activity fs-5" aria-hidden="true"></i>
          </div>
          <div>
            <h2 id="earthquakes-heading" className="h6 mb-0 text-dark fw-bold">
              {t('earthquakeTitle')}
            </h2>
            <span className="text-secondary" style={{ fontSize: '0.75rem' }}>
              {t('earthquakeSourceLabel')}
            </span>
          </div>
        </div>
        {items.length > 0 && (
          <span className="badge bg-light text-dark border px-2 py-1 rounded-pill small">
            {t('earthquakeEventCount', { count: items.length })}
          </span>
        )}
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && items.length === 0 && (
        <AsyncEmpty
          icon="bi-activity"
          variant="safe"
          title={t('emptyEarthquakesTitle')}
          message={t('emptyEarthquakes')}
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className={`d-flex flex-column gap-3 ${items.length > 3 ? 'ks-scrollable-list' : ''}`}>
          {items.map((eq, idx) => {
            const hasCoords = typeof eq.latitude === 'number' && typeof eq.longitude === 'number'
            const isSignificant = typeof eq.magnitude === 'number' && eq.magnitude >= 5.0
            const rawDateTime = eq.date_time || eq.dateTime || ''
            const eventDateTime = formatEarthquakeDateTime(rawDateTime, locale)

            return (
              <div
                key={eq.id || idx}
                className="ks-earthquake-card"
              >
                <div className="d-flex flex-wrap align-items-start justify-content-between gap-2 mb-3">
                  <div className="ks-earthquake-time">
                    <span className="ks-earthquake-time-icon" aria-hidden="true">
                      <i className="bi bi-calendar2-check"></i>
                    </span>
                    <div>
                      <span className="ks-earthquake-time-label">{t('earthquakeTimeLabel')}</span>
                      <time className="ks-earthquake-time-value" dateTime={rawDateTime || undefined}>
                        <span className="ks-earthquake-time-date">{eventDateTime.date}</span>
                        {eventDateTime.time && (
                          <span className="ks-earthquake-time-clock">
                            {eventDateTime.time} <span className="ks-earthquake-timezone">{t('earthquakeTimeZone')}</span>
                          </span>
                        )}
                      </time>
                    </div>
                  </div>
                  <div className="d-flex align-items-center gap-2">
                    {typeof eq.magnitude === 'number' && (
                      <span
                        className={`badge ks-magnitude-badge ${
                          isSignificant ? 'bg-danger text-white' : 'bg-warning text-dark'
                        }`}
                      >
                        M {eq.magnitude.toFixed(1)}
                      </span>
                    )}
                    {typeof eq.depth_km === 'number' && (
                      <span className="badge bg-light text-dark border small">
                        <i className="bi bi-arrow-down me-1 text-muted"></i>
                        {eq.depth_km} km
                      </span>
                    )}
                  </div>
                </div>

                {eq.region && (
                  <div className="mb-2">
                    <span className="small text-muted d-block">{t('region')}:</span>
                    <span className="small fw-semibold text-dark lh-sm">{eq.region}</span>
                  </div>
                )}

                <div className="d-flex flex-wrap align-items-center gap-2 pt-2 border-top border-light-subtle small">
                  {hasCoords && (
                    <span className="badge bg-light text-secondary border">
                      <i className="bi bi-crosshair me-1 text-muted"></i>
                      {eq.latitude}, {eq.longitude}
                    </span>
                  )}

                  {eq.potential && (
                    <span
                      className={`badge rounded-pill ${
                        eq.potential.toLowerCase().includes('tidak')
                          ? 'bg-success-subtle text-success-emphasis border border-success-subtle'
                          : 'bg-danger-subtle text-danger-emphasis border border-danger-subtle'
                      }`}
                    >
                      <i
                        className={`bi ${
                          eq.potential.toLowerCase().includes('tidak')
                            ? 'bi-shield-check'
                            : 'bi-exclamation-triangle-fill'
                        } me-1`}
                      ></i>
                      {eq.potential}
                    </span>
                  )}

                  {eq.felt_areas && (
                    <div className="w-100 mt-1 small text-secondary">
                      <i className="bi bi-soundwave me-1 text-primary"></i>
                      <span>{t('felt')}: </span>
                      <strong className="text-dark">{eq.felt_areas}</strong>
                    </div>
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
