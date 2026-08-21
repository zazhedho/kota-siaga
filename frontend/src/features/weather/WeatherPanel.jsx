import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

function getWeatherIcon(desc) {
  const norm = (desc || '').toLowerCase()
  if (norm.includes('petir') || norm.includes('thunder') || norm.includes('kilat')) {
    return 'bi-cloud-lightning-rain-fill text-danger'
  }
  if (norm.includes('hujan lebat') || norm.includes('heavy rain')) {
    return 'bi-cloud-rain-heavy-fill text-primary'
  }
  if (norm.includes('hujan') || norm.includes('rain') || norm.includes('drizzle')) {
    return 'bi-cloud-rain-fill text-primary'
  }
  if (norm.includes('cerah berawan') || norm.includes('partly cloudy')) {
    return 'bi-cloud-sun-fill text-warning'
  }
  if (norm.includes('cerah') || norm.includes('clear') || norm.includes('sunny')) {
    return 'bi-sun-fill text-warning'
  }
  if (norm.includes('kabut') || norm.includes('fog') || norm.includes('mist')) {
    return 'bi-cloud-fog2-fill text-secondary'
  }
  return 'bi-clouds-fill text-secondary'
}

export function WeatherPanel({ items = [], loading = false, error = null, onRetry }) {
  const { locale, t } = useLocale()

  return (
    <div className="ks-card" aria-labelledby="weather-heading">
      <div className="d-flex align-items-center justify-content-between ks-card-header">
        <div className="d-flex align-items-center gap-2">
          <div className="p-2 rounded-circle bg-info-subtle text-info-emphasis d-inline-flex">
            <i className="bi bi-cloud-sun fs-5" aria-hidden="true"></i>
          </div>
          <div>
            <h2 id="weather-heading" className="h6 mb-0 text-dark fw-bold">
              {t('weatherTitle')}
            </h2>
            <span className="text-secondary" style={{ fontSize: '0.75rem' }}>
              {t('weatherSourceLabel')}
            </span>
          </div>
        </div>
        {items.length > 0 && (
          <span className="badge bg-light text-dark border px-2 py-1 rounded-pill small">
            {t('forecastPeriodCount', { count: items.length })}
          </span>
        )}
      </div>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && items.length === 0 && (
        <AsyncEmpty
          icon="bi-cloud-sun"
          variant="info"
          title={t('emptyWeatherTitle')}
          message={t('emptyWeather')}
        />
      )}

      {!loading && !error && items.length > 0 && (
        <div className={`d-flex flex-column gap-2 ${items.length > 5 ? 'ks-scrollable-list' : ''}`}>
          {items.map((item, idx) => {
            const desc =
              locale === 'en' && item.weather_description_en
                ? item.weather_description_en
                : item.weather_description || item.weather || '-'

            const timeLabel = item.local_datetime || item.datetime || ''
            const weatherIcon = getWeatherIcon(desc)

            return (
              <div key={item.id || idx} className="ks-forecast-item">
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-2">
                  <div className="d-flex align-items-center gap-2">
                    <i className={`bi ${weatherIcon} fs-5`} aria-hidden="true"></i>
                    <span className="fw-bold text-dark small">{timeLabel}</span>
                  </div>
                  <span className="badge bg-light text-dark border px-2 py-1 rounded-pill small">
                    {desc}
                  </span>
                </div>

                <div className="d-flex flex-wrap align-items-center justify-content-between gap-3 pt-2 border-top border-light-subtle">
                  {typeof item.temperature_c === 'number' && (
                    <div className="d-flex align-items-baseline gap-1">
                      <span className="ks-temp-badge">{item.temperature_c}°C</span>
                    </div>
                  )}

                  <div className="d-flex flex-wrap align-items-center gap-2 small text-secondary">
                    {typeof item.humidity_percent === 'number' && (
                      <span className="badge bg-light text-secondary border d-inline-flex align-items-center gap-1">
                        <i className="bi bi-droplet-fill text-info" aria-hidden="true"></i>
                        <span>{t('humidity')}: <strong>{item.humidity_percent}%</strong></span>
                      </span>
                    )}
                    {typeof item.precipitation_mm === 'number' && (
                      <span className="badge bg-light text-secondary border d-inline-flex align-items-center gap-1">
                        <i className="bi bi-cloud-rain text-primary" aria-hidden="true"></i>
                        <span>{t('precipitation')}: <strong>{item.precipitation_mm} mm</strong></span>
                      </span>
                    )}
                    {(item.wind_speed !== undefined || item.wind_direction) && (
                      <span className="badge bg-light text-secondary border d-inline-flex align-items-center gap-1">
                        <i className="bi bi-wind text-secondary" aria-hidden="true"></i>
                        <span>
                          {t('wind')}: <strong>{item.wind_speed ? `${item.wind_speed} km/h` : ''} {item.wind_direction || ''}</strong>
                        </span>
                      </span>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
