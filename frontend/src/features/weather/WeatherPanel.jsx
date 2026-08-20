import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

export function WeatherPanel({ items = [], loading = false, error = null, onRetry }) {
  const { locale, t } = useLocale()

  return (
    <div className="ks-card h-100" aria-labelledby="weather-heading">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h2 id="weather-heading" className="h5 mb-0 text-dark fw-bold">
          <i className="bi bi-cloud-sun me-2 text-primary"></i>
          {t('weatherTitle')}
        </h2>
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && items.length === 0 && (
        <AsyncEmpty message={t('emptyWeather')} />
      )}

      {!loading && !error && items.length > 0 && (
        <div className="d-flex flex-column gap-2">
          {items.map((item, idx) => {
            const desc =
              locale === 'en' && item.weather_description_en
                ? item.weather_description_en
                : item.weather_description || item.weather || '-'

            const timeLabel = item.local_datetime || item.datetime || ''

            return (
              <div key={item.id || idx} className="ks-forecast-item">
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-1">
                  <span className="fw-semibold text-dark small">{timeLabel}</span>
                  <span className="badge bg-light text-dark border">{desc}</span>
                </div>
                <div className="row g-2 small text-secondary">
                  {typeof item.temperature_c === 'number' && (
                    <div className="col-6 col-sm-3">
                      <span>{t('temperature')}: </span>
                      <strong className="text-dark">{item.temperature_c}°C</strong>
                    </div>
                  )}
                  {typeof item.humidity_percent === 'number' && (
                    <div className="col-6 col-sm-3">
                      <span>{t('humidity')}: </span>
                      <strong className="text-dark">{item.humidity_percent}%</strong>
                    </div>
                  )}
                  {typeof item.precipitation_mm === 'number' && (
                    <div className="col-6 col-sm-3">
                      <span>{t('precipitation')}: </span>
                      <strong className="text-dark">{item.precipitation_mm} mm</strong>
                    </div>
                  )}
                  {(item.wind_speed !== undefined || item.wind_direction) && (
                    <div className="col-6 col-sm-3">
                      <span>{t('wind')}: </span>
                      <strong className="text-dark">
                        {item.wind_speed ? `${item.wind_speed} km/h` : ''} {item.wind_direction || ''}
                      </strong>
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
