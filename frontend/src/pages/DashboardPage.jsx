import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocale } from '../shared/i18n'
import { LocationSelector } from '../features/location/LocationSelector'
import { WeatherPanel } from '../features/weather/WeatherPanel'
import { WarningPanel } from '../features/warning/WarningPanel'
import { EarthquakePanel } from '../features/earthquake/EarthquakePanel'
import { HospitalPanel } from '../features/hospital/HospitalPanel'
import { getForecast } from '../features/weather/weatherService'
import { listWarnings } from '../features/warning/warningService'
import { listLatest } from '../features/earthquake/earthquakeService'
import { listHospitals } from '../features/hospital/hospitalService'
import { readStoredLocation, writeStoredLocation, clearStoredLocation } from '../shared/storage'
import { getApiErrorMessage } from '../shared/api/client'

export function DashboardPage() {
  const { t } = useLocale()
  const [storedLocation] = useState(() => readStoredLocation())
  const [location, setLocation] = useState(null)

  // Weather state
  const [weatherData, setWeatherData] = useState([])
  const [loadingWeather, setLoadingWeather] = useState(false)
  const [errorWeather, setErrorWeather] = useState(null)

  // Warning state
  const [warningData, setWarningData] = useState([])
  const [loadingWarnings, setLoadingWarnings] = useState(false)
  const [errorWarnings, setErrorWarnings] = useState(null)

  // Earthquake state
  const [earthquakeData, setEarthquakeData] = useState([])
  const [loadingEarthquakes, setLoadingEarthquakes] = useState(false)
  const [errorEarthquakes, setErrorEarthquakes] = useState(null)

  // Hospital state
  const [hospitalData, setHospitalData] = useState([])
  const [hospitalTotal, setHospitalTotal] = useState(0)
  const [hospitalPage, setHospitalPage] = useState(1)
  const [hasNextHospitalPage, setHasNextHospitalPage] = useState(false)
  const [loadingHospitals, setLoadingHospitals] = useState(false)
  const [loadingMoreHospitals, setLoadingMoreHospitals] = useState(false)
  const [errorHospitals, setErrorHospitals] = useState(null)
  const [hospitalSearch, setHospitalSearch] = useState('')

  const abortControllerRef = useRef(null)
  const hospitalRequestIdRef = useRef(0)

  const fetchWeather = useCallback(async (villageCode, signal) => {
    setLoadingWeather(true)
    setErrorWeather(null)
    try {
      const items = await getForecast(villageCode, signal)
      setWeatherData(items)
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorWeather(err.message || t('genericError'))
      }
    } finally {
      setLoadingWeather(false)
    }
  }, [t])

  const fetchWarnings = useCallback(async (provinceName, signal) => {
    setLoadingWarnings(true)
    setErrorWarnings(null)
    try {
      const items = await listWarnings(provinceName, signal)
      setWarningData(items)
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorWarnings(err.message || t('genericError'))
      }
    } finally {
      setLoadingWarnings(false)
    }
  }, [t])

  const fetchEarthquakes = useCallback(async (signal) => {
    setLoadingEarthquakes(true)
    setErrorEarthquakes(null)
    try {
      const items = await listLatest(signal)
      setEarthquakeData(items)
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorEarthquakes(err.message || t('genericError'))
      }
    } finally {
      setLoadingEarthquakes(false)
    }
  }, [t])

  const fetchHospitals = useCallback(async (cityId, page = 1, append = false, signal, search = '') => {
    const requestId = ++hospitalRequestIdRef.current
    const normalizedSearch = typeof search === 'string' ? search.trim() : ''

    if (append) {
      setLoadingMoreHospitals(true)
    } else {
      setLoadingHospitals(true)
      setLoadingMoreHospitals(false)
      setErrorHospitals(null)
    }

    try {
      const result = await listHospitals(cityId, page, normalizedSearch, signal)
      if (requestId !== hospitalRequestIdRef.current) return

      if (append) {
        setHospitalData((prev) => [...prev, ...(result.rows || [])])
      } else {
        setHospitalData(result.rows || [])
      }
      setHospitalTotal(result.total)
      setHospitalPage(result.page)
      setHasNextHospitalPage(result.nextPage)
    } catch (err) {
      if (requestId === hospitalRequestIdRef.current && err.name !== 'AbortError') {
        setErrorHospitals(getApiErrorMessage(err, t) || t('genericError'))
      }
    } finally {
      if (requestId === hospitalRequestIdRef.current) {
        setLoadingHospitals(false)
        setLoadingMoreHospitals(false)
      }
    }
  }, [t])

  const handleHospitalSearch = useCallback((search) => {
    const normalizedSearch = typeof search === 'string' ? search.trim() : ''
    setHospitalSearch(normalizedSearch)

    if (location) {
      fetchHospitals(
        location.city.id,
        1,
        false,
        abortControllerRef.current?.signal,
        normalizedSearch,
      )
    }
  }, [fetchHospitals, location])

  const loadAllFeatures = useCallback((loc) => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    const controller = new AbortController()
    abortControllerRef.current = controller

    const villageCode = loc.village.code || loc.adm4
    const provinceName = loc.province.name
    const cityId = loc.city.id

    // Independent feature execution
    fetchWeather(villageCode, controller.signal)
    fetchWarnings(provinceName, controller.signal)
    fetchEarthquakes(controller.signal)
    fetchHospitals(cityId, 1, false, controller.signal, '')
  }, [fetchWeather, fetchWarnings, fetchEarthquakes, fetchHospitals])

  const handleLocationComplete = useCallback((selectedLoc) => {
    setLocation(selectedLoc)
    hospitalRequestIdRef.current += 1
    setHospitalSearch('')
    setHospitalData([])
    setHospitalTotal(0)
    setHospitalPage(1)
    setHasNextHospitalPage(false)
    setLoadingHospitals(false)
    setLoadingMoreHospitals(false)
    setErrorHospitals(null)

    if (!selectedLoc) {
      clearStoredLocation()
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }
      setWeatherData([])
      setWarningData([])
      setEarthquakeData([])
      return
    }

    writeStoredLocation(selectedLoc)
    loadAllFeatures(selectedLoc)
  }, [loadAllFeatures])

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }
    }
  }, [])

  return (
    <div className="d-flex flex-column gap-4">
      {/* Cascading Location Controls */}
      <LocationSelector
        initialLocation={storedLocation}
        onComplete={handleLocationComplete}
      />

      {/* When no location is selected yet: Modern Welcome Onboarding Hero */}
      {!location && (
        <div className="ks-welcome-hero">
          <div className="text-center mb-4">
            <span className="badge bg-primary-subtle text-primary border border-primary-subtle px-3 py-1 rounded-pill small fw-bold mb-2">
              <i className="bi bi-shield-fill-check me-2"></i>
              {t('welcomeBadge')}
            </span>
            <h2 className="h4 fw-bold text-dark mb-2">{t('welcomeTitle')}</h2>
            <p className="text-secondary mx-auto mb-0" style={{ maxWidth: '640px' }}>
              {t('emptyDashboardPrompt')}
            </p>
          </div>

          <div className="row g-3 pt-2">
            <div className="col-12 col-sm-6 col-lg-3">
              <div className="ks-feature-preview-card">
                <div className="p-2 rounded-circle bg-info-subtle text-info d-inline-flex mb-2">
                  <i className="bi bi-cloud-sun fs-5"></i>
                </div>
                <h3 className="h6 fw-bold text-dark mb-1">{t('weatherTitle')}</h3>
                <p className="small text-secondary mb-0">{t('featureWeatherDesc')}</p>
              </div>
            </div>

            <div className="col-12 col-sm-6 col-lg-3">
              <div className="ks-feature-preview-card">
                <div className="p-2 rounded-circle bg-warning-subtle text-warning d-inline-flex mb-2">
                  <i className="bi bi-shield-exclamation fs-5"></i>
                </div>
                <h3 className="h6 fw-bold text-dark mb-1">{t('warningTitle')}</h3>
                <p className="small text-secondary mb-0">{t('featureWarningDesc')}</p>
              </div>
            </div>

            <div className="col-12 col-sm-6 col-lg-3">
              <div className="ks-feature-preview-card">
                <div className="p-2 rounded-circle bg-danger-subtle text-danger d-inline-flex mb-2">
                  <i className="bi bi-activity fs-5"></i>
                </div>
                <h3 className="h6 fw-bold text-dark mb-1">{t('earthquakeTitle')}</h3>
                <p className="small text-secondary mb-0">{t('featureEarthquakeDesc')}</p>
              </div>
            </div>

            <div className="col-12 col-sm-6 col-lg-3">
              <div className="ks-feature-preview-card">
                <div className="p-2 rounded-circle bg-primary-subtle text-primary d-inline-flex mb-2">
                  <i className="bi bi-hospital fs-5"></i>
                </div>
                <h3 className="h6 fw-bold text-dark mb-1">{t('hospitalTitle')}</h3>
                <p className="small text-secondary mb-0">{t('featureHospitalDesc')}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Monitored Location Summary & Feature Grid */}
      {location && (
        <>
          <div
            className="ks-selected-location-banner d-flex flex-wrap align-items-center justify-content-between gap-3 shadow-xs"
            role="status"
            aria-live="polite"
          >
            <div className="d-flex align-items-center gap-3">
              <div className="p-2 rounded-circle bg-primary text-white d-inline-flex shadow-xs">
                <i className="bi bi-geo-alt-fill fs-5" aria-hidden="true"></i>
              </div>
              <div>
                <span className="small text-primary fw-bold text-uppercase d-block" style={{ fontSize: '0.75rem', letterSpacing: '0.04em' }}>
                  {t('selectedLocationLabel')}
                </span>
                <strong className="text-dark fs-6">
                  {location.village.name}, {location.district.name}, {location.city.name}, {location.province.name}
                </strong>
              </div>
            </div>
            <span className="badge bg-white text-primary border border-primary-subtle px-3 py-2 rounded-pill font-monospace small shadow-xs">
              ADM4: {location.village.code || location.adm4}
            </span>
          </div>

          <div className="row g-4 align-items-start">
            {/* Left Column: Weather and Hospitals (below alerts on small screens) */}
            <div className="col-12 col-lg-7 order-2 order-lg-1 d-flex flex-column gap-4">
              <WeatherPanel
                items={weatherData}
                loading={loadingWeather}
                error={errorWeather}
                onRetry={() => location && fetchWeather(location.village.code || location.adm4)}
              />
              <HospitalPanel
                key={location.village.code || location.adm4}
                hospitals={hospitalData}
                total={hospitalTotal}
                loading={loadingHospitals}
                loadingMore={loadingMoreHospitals}
                error={errorHospitals}
                hasNextPage={hasNextHospitalPage}
                search={hospitalSearch}
                onSearch={handleHospitalSearch}
                onLoadMore={() =>
                  location &&
                  fetchHospitals(
                    location.city.id,
                    hospitalPage + 1,
                    true,
                    abortControllerRef.current?.signal,
                    hospitalSearch,
                  )
                }
                onRetry={() =>
                  location &&
                  fetchHospitals(
                    location.city.id,
                    1,
                    false,
                    abortControllerRef.current?.signal,
                    hospitalSearch,
                  )
                }
              />
            </div>

            {/* Right Column: Warnings and Earthquakes (first on small screens) */}
            <div className="col-12 col-lg-5 order-1 order-lg-2 d-flex flex-column gap-4">
              <WarningPanel
                warnings={warningData}
                loading={loadingWarnings}
                error={errorWarnings}
                onRetry={() => location && fetchWarnings(location.province.name)}
              />
              <EarthquakePanel
                items={earthquakeData}
                loading={loadingEarthquakes}
                error={errorEarthquakes}
                onRetry={() => fetchEarthquakes()}
              />
            </div>
          </div>
        </>
      )}
    </div>
  )
}
