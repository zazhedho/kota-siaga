import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocale } from '../../shared/i18n'
import { getApiErrorMessage } from '../../shared/api/client'
import {
  listProvinces,
  listCities,
  listDistricts,
  listVillages,
} from './locationService'

export function LocationSelector({ onComplete, initialLocation = null }) {
  const { t } = useLocale()

  const [provinces, setProvinces] = useState([])
  const [cities, setCities] = useState([])
  const [districts, setDistricts] = useState([])
  const [villages, setVillages] = useState([])

  const [selectedProvinceId, setSelectedProvinceId] = useState('')
  const [selectedCityId, setSelectedCityId] = useState('')
  const [selectedDistrictId, setSelectedDistrictId] = useState('')
  const [selectedVillageId, setSelectedVillageId] = useState('')

  const [loadingProvince, setLoadingProvince] = useState(false)
  const [loadingCity, setLoadingCity] = useState(false)
  const [loadingDistrict, setLoadingDistrict] = useState(false)
  const [loadingVillage, setLoadingVillage] = useState(false)

  const [errorProvince, setErrorProvince] = useState(null)
  const [errorCity, setErrorCity] = useState(null)
  const [errorDistrict, setErrorDistrict] = useState(null)
  const [errorVillage, setErrorVillage] = useState(null)

  const provinceAbortRef = useRef(null)
  const cityAbortRef = useRef(null)
  const districtAbortRef = useRef(null)
  const villageAbortRef = useRef(null)
  const restoredLocationRef = useRef('')

  // Load provinces on mount
  const fetchProvinces = useCallback(async () => {
    if (provinceAbortRef.current) provinceAbortRef.current.abort()
    const controller = new AbortController()
    provinceAbortRef.current = controller

    setLoadingProvince(true)
    setErrorProvince(null)

    try {
      const result = await listProvinces(controller.signal)
      setProvinces(result.rows || [])
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorProvince(err)
      }
    } finally {
      setLoadingProvince(false)
    }
  }, [])

  useEffect(() => {
    fetchProvinces()
    return () => {
      if (provinceAbortRef.current) provinceAbortRef.current.abort()
      if (cityAbortRef.current) cityAbortRef.current.abort()
      if (districtAbortRef.current) districtAbortRef.current.abort()
      if (villageAbortRef.current) villageAbortRef.current.abort()
    }
  }, [fetchProvinces])

  // Load cities when province changes
  const fetchCities = useCallback(async (provinceId) => {
    if (cityAbortRef.current) cityAbortRef.current.abort()
    const controller = new AbortController()
    cityAbortRef.current = controller

    setLoadingCity(true)
    setErrorCity(null)

    try {
      const result = await listCities(provinceId, controller.signal)
      setCities(result.rows || [])
      return result
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorCity(err)
      }
    } finally {
      setLoadingCity(false)
    }
  }, [])

  // Load districts when city changes
  const fetchDistricts = useCallback(async (cityId) => {
    if (districtAbortRef.current) districtAbortRef.current.abort()
    const controller = new AbortController()
    districtAbortRef.current = controller

    setLoadingDistrict(true)
    setErrorDistrict(null)

    try {
      const result = await listDistricts(cityId, controller.signal)
      setDistricts(result.rows || [])
      return result
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorDistrict(err)
      }
    } finally {
      setLoadingDistrict(false)
    }
  }, [])

  // Load villages when district changes
  const fetchVillages = useCallback(async (districtId) => {
    if (villageAbortRef.current) villageAbortRef.current.abort()
    const controller = new AbortController()
    villageAbortRef.current = controller

    setLoadingVillage(true)
    setErrorVillage(null)

    try {
      const result = await listVillages(districtId, controller.signal)
      setVillages(result.rows || [])
      return result
    } catch (err) {
      if (err?.name !== 'AbortError') {
        setErrorVillage(err)
      }
    } finally {
      setLoadingVillage(false)
    }
  }, [])

  // Restore initial location if provided
  useEffect(() => {
    if (!initialLocation || provinces.length === 0) return

    const { provinceId, cityId, districtId, villageId } = initialLocation
    const restoreKey = [provinceId, cityId, districtId, villageId].join(':')
    if (restoredLocationRef.current === restoreKey) return

    const province = provinces.find((p) => String(p.id) === String(provinceId))
    if (!province) return

    restoredLocationRef.current = restoreKey
    let cancelled = false

    const restore = async () => {
      setSelectedProvinceId(String(province.id))

      const cityResult = await fetchCities(province.id)
      const city = cityResult?.rows?.find((item) => String(item.id) === String(cityId))
      if (cancelled || !city) return
      setSelectedCityId(String(city.id))

      const districtResult = await fetchDistricts(city.id)
      const district = districtResult?.rows?.find((item) => String(item.id) === String(districtId))
      if (cancelled || !district) return
      setSelectedDistrictId(String(district.id))

      const villageResult = await fetchVillages(district.id)
      const village = villageResult?.rows?.find((item) => String(item.id) === String(villageId))
      if (cancelled || !village) return
      setSelectedVillageId(String(village.id))

      onComplete?.({
        province,
        city,
        district,
        village,
        adm4: village.code || initialLocation.adm4,
      })
    }

    restore()
    return () => {
      cancelled = true
    }
  }, [initialLocation, provinces, fetchCities, fetchDistricts, fetchVillages, onComplete])

  const handleProvinceChange = (e) => {
    const pId = e.target.value
    setSelectedProvinceId(pId)

    // Clear descendants
    setSelectedCityId('')
    setSelectedDistrictId('')
    setSelectedVillageId('')
    setCities([])
    setDistricts([])
    setVillages([])
    setErrorCity(null)
    setErrorDistrict(null)
    setErrorVillage(null)
    onComplete?.(null)

    if (pId) {
      fetchCities(pId)
    }
  }

  const handleCityChange = (e) => {
    const cId = e.target.value
    setSelectedCityId(cId)

    // Clear descendants
    setSelectedDistrictId('')
    setSelectedVillageId('')
    setDistricts([])
    setVillages([])
    setErrorDistrict(null)
    setErrorVillage(null)
    onComplete?.(null)

    if (cId) {
      fetchDistricts(cId)
    }
  }

  const handleDistrictChange = (e) => {
    const dId = e.target.value
    setSelectedDistrictId(dId)

    // Clear descendants
    setSelectedVillageId('')
    setVillages([])
    setErrorVillage(null)
    onComplete?.(null)

    if (dId) {
      fetchVillages(dId)
    }
  }

  const handleVillageChange = (e) => {
    const vId = e.target.value
    setSelectedVillageId(vId)

    if (!vId) {
      onComplete?.(null)
      return
    }

    const province = provinces.find((p) => String(p.id) === String(selectedProvinceId))
    const city = cities.find((c) => String(c.id) === String(selectedCityId))
    const district = districts.find((d) => String(d.id) === String(selectedDistrictId))
    const village = villages.find((v) => String(v.id) === String(vId))

    if (province && city && district && village) {
      onComplete?.({
        province,
        city,
        district,
        village,
        adm4: village.code,
      })
    }
  }

  return (
    <div className="ks-card mb-4" aria-labelledby="location-heading">
      <div className="mb-3">
        <h2 id="location-heading" className="h5 mb-1 text-dark fw-bold">
          <i className="bi bi-geo-alt me-2 text-primary"></i>
          {t('locationSectionTitle')}
        </h2>
        <p className="text-secondary small mb-0">{t('locationInstruction')}</p>
      </div>

      <div className="row g-3">
        {/* Province */}
        <div className="col-12 col-md-6 col-lg-3">
          <label htmlFor="province-select" className="form-label fw-semibold small">
            {t('province')}
          </label>
          <select
            id="province-select"
            className="form-select"
            value={selectedProvinceId}
            onChange={handleProvinceChange}
            disabled={loadingProvince}
            aria-busy={loadingProvince}
          >
            <option value="">
              {loadingProvince ? t('loadingLocations') : t('selectProvince')}
            </option>
            {provinces.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          {errorProvince && (
            <div className="mt-1 small text-danger d-flex align-items-center justify-content-between">
              <span>{getApiErrorMessage(errorProvince, t)}</span>
              <button
                type="button"
                className="btn btn-link btn-sm p-0 ms-2 text-decoration-none"
                onClick={fetchProvinces}
              >
                {t('retry')}
              </button>
            </div>
          )}
        </div>

        {/* City / Regency */}
        <div className="col-12 col-md-6 col-lg-3">
          <label htmlFor="city-select" className="form-label fw-semibold small">
            {t('city')}
          </label>
          <select
            id="city-select"
            className="form-select"
            value={selectedCityId}
            onChange={handleCityChange}
            disabled={!selectedProvinceId || loadingCity}
            aria-busy={loadingCity}
          >
            <option value="">
              {loadingCity ? t('loadingLocations') : t('selectCity')}
            </option>
            {cities.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          {errorCity && (
            <div className="mt-1 small text-danger d-flex align-items-center justify-content-between">
              <span>{getApiErrorMessage(errorCity, t)}</span>
              <button
                type="button"
                className="btn btn-link btn-sm p-0 ms-2 text-decoration-none"
                onClick={() => fetchCities(selectedProvinceId)}
              >
                {t('retry')}
              </button>
            </div>
          )}
        </div>

        {/* District */}
        <div className="col-12 col-md-6 col-lg-3">
          <label htmlFor="district-select" className="form-label fw-semibold small">
            {t('district')}
          </label>
          <select
            id="district-select"
            className="form-select"
            value={selectedDistrictId}
            onChange={handleDistrictChange}
            disabled={!selectedCityId || loadingDistrict}
            aria-busy={loadingDistrict}
          >
            <option value="">
              {loadingDistrict ? t('loadingLocations') : t('selectDistrict')}
            </option>
            {districts.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
          {errorDistrict && (
            <div className="mt-1 small text-danger d-flex align-items-center justify-content-between">
              <span>{getApiErrorMessage(errorDistrict, t)}</span>
              <button
                type="button"
                className="btn btn-link btn-sm p-0 ms-2 text-decoration-none"
                onClick={() => fetchDistricts(selectedCityId)}
              >
                {t('retry')}
              </button>
            </div>
          )}
        </div>

        {/* Village */}
        <div className="col-12 col-md-6 col-lg-3">
          <label htmlFor="village-select" className="form-label fw-semibold small">
            {t('village')}
          </label>
          <select
            id="village-select"
            className="form-select"
            value={selectedVillageId}
            onChange={handleVillageChange}
            disabled={!selectedDistrictId || loadingVillage}
            aria-busy={loadingVillage}
          >
            <option value="">
              {loadingVillage ? t('loadingLocations') : t('selectVillage')}
            </option>
            {villages.map((v) => (
              <option key={v.id} value={v.id}>
                {v.name}
              </option>
            ))}
          </select>
          {errorVillage && (
            <div className="mt-1 small text-danger d-flex align-items-center justify-content-between">
              <span>{getApiErrorMessage(errorVillage, t)}</span>
              <button
                type="button"
                className="btn btn-link btn-sm p-0 ms-2 text-decoration-none"
                onClick={() => fetchVillages(selectedDistrictId)}
              >
                {t('retry')}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
