import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocale } from '../shared/i18n'
import { getApiErrorMessage } from '../shared/api/client'
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

  const abortControllerRef = useRef(null)

  const fetchWeather = useCallback(async (villageCode, signal) => {
    setLoadingWeather(true)
    setErrorWeather(null)
    try {
      const items = await getForecast(villageCode, signal)
      setWeatherData(items)
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorWeather(err)
      }
    } finally {
      setLoadingWeather(false)
    }
  }, [])

  const fetchWarnings = useCallback(async (provinceName, signal) => {
    setLoadingWarnings(true)
    setErrorWarnings(null)
    try {
      const items = await listWarnings(provinceName, signal)
      setWarningData(items)
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorWarnings(err)
      }
    } finally {
      setLoadingWarnings(false)
    }
  }, [])

  const fetchEarthquakes = useCallback(async (signal) => {
    setLoadingEarthquakes(true)
    setErrorEarthquakes(null)
    try {
      const items = await listLatest(signal)
      setEarthquakeData(items)
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorEarthquakes(err)
      }
    } finally {
      setLoadingEarthquakes(false)
    }
  }, [])

  const fetchHospitals = useCallback(async (cityId, page = 1, append = false, signal) => {
    if (append) {
      setLoadingMoreHospitals(true)
    } else {
      setLoadingHospitals(true)
      setErrorHospitals(null)
    }

    try {
      const result = await listHospitals(cityId, page, signal)
      if (append) {
        setHospitalData((prev) => [...prev, ...(result.rows || [])])
      } else {
        setHospitalData(result.rows || [])
      }
      setHospitalTotal(result.total)
      setHospitalPage(result.page)
      setHasNextHospitalPage(result.nextPage)
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorHospitals(err)
      }
    } finally {
      setLoadingHospitals(false)
      setLoadingMoreHospitals(false)
    }
  }, [])

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
    fetchHospitals(cityId, 1, false, controller.signal)
  }, [fetchWeather, fetchWarnings, fetchEarthquakes, fetchHospitals])

  const handleLocationComplete = useCallback((selectedLoc) => {
    setLocation(selectedLoc)

    if (!selectedLoc) {
      clearStoredLocation()
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }
      setWeatherData([])
      setWarningData([])
      setEarthquakeData([])
      setHospitalData([])
      setHospitalTotal(0)
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

      {/* When no location is selected yet */}
      {!location && (
        <div className="ks-card text-center py-5">
          <i className="bi bi-geo-alt-fill fs-1 text-primary mb-3 d-block" aria-hidden="true"></i>
          <h2 className="h5 fw-bold text-dark mb-2">{t('locationSectionTitle')}</h2>
          <p className="text-secondary mx-auto mb-0" style={{ maxWidth: '580px' }}>
            {t('emptyDashboardPrompt')}
          </p>
        </div>
      )}

      {/* Monitored Location Summary & Feature Grid */}
      {location && (
        <>
          <div className="p-3 bg-primary-subtle border border-primary-subtle rounded d-flex flex-wrap align-items-center justify-content-between gap-2">
            <div className="d-flex align-items-center gap-2">
              <i className="bi bi-pin-map-fill text-primary fs-5" aria-hidden="true"></i>
              <div>
                <span className="small text-secondary text-uppercase fw-semibold d-block">
                  {t('selectedLocationLabel')}
                </span>
                <strong className="text-dark">
                  {location.village.name}, {location.district.name}, {location.city.name}, {location.province.name}
                </strong>
              </div>
            </div>
            <span className="badge bg-primary">
              ADM4: {location.village.code || location.adm4}
            </span>
          </div>

          <div className="row g-4">
            {/* Weather & Warnings Row */}
            <div className="col-12 col-lg-7">
              <WeatherPanel
                items={weatherData}
                loading={loadingWeather}
                error={getApiErrorMessage(errorWeather, t)}
                onRetry={() => location && fetchWeather(location.village.code || location.adm4)}
              />
            </div>
            <div className="col-12 col-lg-5">
              <WarningPanel
                warnings={warningData}
                loading={loadingWarnings}
                error={getApiErrorMessage(errorWarnings, t)}
                onRetry={() => location && fetchWarnings(location.province.name)}
              />
            </div>

            {/* Earthquakes & Hospitals Row */}
            <div className="col-12 col-lg-5">
              <EarthquakePanel
                items={earthquakeData}
                loading={loadingEarthquakes}
                error={getApiErrorMessage(errorEarthquakes, t)}
                onRetry={() => fetchEarthquakes()}
              />
            </div>
            <div className="col-12 col-lg-7">
              <HospitalPanel
                hospitals={hospitalData}
                total={hospitalTotal}
                loading={loadingHospitals}
                loadingMore={loadingMoreHospitals}
                error={getApiErrorMessage(errorHospitals, t)}
                hasNextPage={hasNextHospitalPage}
                onLoadMore={() =>
                  location &&
                  fetchHospitals(location.city.id, hospitalPage + 1, true)
                }
                onRetry={() =>
                  location && fetchHospitals(location.city.id, 1, false)
                }
              />
            </div>
          </div>
        </>
      )}
    </div>
  )
}
