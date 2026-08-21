import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocale } from '../../shared/i18n'
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
  const restoredRef = useRef(false)

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
      if (err.name !== 'AbortError') {
        setErrorProvince(err.message || t('genericError'))
      }
    } finally {
      setLoadingProvince(false)
    }
  }, [t])

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
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorCity(err.message || t('genericError'))
      }
    } finally {
      setLoadingCity(false)
    }
  }, [t])

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
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorDistrict(err.message || t('genericError'))
      }
    } finally {
      setLoadingDistrict(false)
    }
  }, [t])

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
    } catch (err) {
      if (err.name !== 'AbortError') {
        setErrorVillage(err.message || t('genericError'))
      }
    } finally {
      setLoadingVillage(false)
    }
  }, [t])

  // Restore initial location if provided
  useEffect(() => {
    if (!initialLocation || provinces.length === 0 || restoredRef.current) return

    const { provinceId, cityId, districtId, villageId } = initialLocation
    const prov = provinces.find((p) => String(p.id) === String(provinceId))
    if (!prov) return

    restoredRef.current = true
    setSelectedProvinceId(String(provinceId))

    listCities(provinceId).then((cRes) => {
      const cRows = cRes.rows || []
      setCities(cRows)
      setSelectedCityId(String(cityId))
      const city = cRows.find((c) => String(c.id) === String(cityId))

      listDistricts(cityId).then((dRes) => {
        const dRows = dRes.rows || []
        setDistricts(dRows)
        setSelectedDistrictId(String(districtId))
        const district = dRows.find((d) => String(d.id) === String(districtId))

        listVillages(districtId).then((vRes) => {
          const vRows = vRes.rows || []
          setVillages(vRows)
          setSelectedVillageId(String(villageId))
          const village = vRows.find((v) => String(v.id) === String(villageId))

          if (prov && city && district && village) {
            onComplete?.({
              province: prov,
              city,
              district,
              village,
              adm4: village.code || initialLocation.adm4,
            })
          }
        })
      })
    })
  }, [initialLocation, provinces, onComplete])

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
    <div className="ks-card ks-location-card mb-4" aria-labelledby="location-heading">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <div className="d-flex align-items-center gap-2">
          <div className="ks-icon-tile bg-primary-subtle text-primary">
            <i className="bi bi-geo-alt-fill" aria-hidden="true"></i>
          </div>
          <div>
            <h2 id="location-heading" className="h6 mb-0 text-dark fw-bold">
              {t('locationSectionTitle')}
            </h2>
            <p className="text-secondary small mb-0">{t('locationInstruction')}</p>
          </div>
        </div>
      </div>

      <div className="row g-3">
        {/* Province */}
        <div className="col-12 col-md-6 col-lg-3">
          <label htmlFor="province-select" className="form-label fw-semibold small text-dark d-flex align-items-center justify-content-between mb-1.5">
            <span>
              <span className="ks-step-badge me-2">1</span>
              {t('province')}
            </span>
            {loadingProvince && (
              <span className="spinner-border spinner-border-sm text-primary" role="status" aria-hidden="true" />
            )}
          </label>
          <select
            id="province-select"
            className="form-select shadow-xs"
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
              <span>{errorProvince}</span>
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
          <label htmlFor="city-select" className="form-label fw-semibold small text-dark d-flex align-items-center justify-content-between mb-1.5">
            <span>
              <span className="ks-step-badge me-2">2</span>
              {t('city')}
            </span>
            {loadingCity && (
              <span className="spinner-border spinner-border-sm text-primary" role="status" aria-hidden="true" />
            )}
          </label>
          <select
            id="city-select"
            className="form-select shadow-xs"
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
              <span>{errorCity}</span>
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
          <label htmlFor="district-select" className="form-label fw-semibold small text-dark d-flex align-items-center justify-content-between mb-1.5">
            <span>
              <span className="ks-step-badge me-2">3</span>
              {t('district')}
            </span>
            {loadingDistrict && (
              <span className="spinner-border spinner-border-sm text-primary" role="status" aria-hidden="true" />
            )}
          </label>
          <select
            id="district-select"
            className="form-select shadow-xs"
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
              <span>{errorDistrict}</span>
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
          <label htmlFor="village-select" className="form-label fw-semibold small text-dark d-flex align-items-center justify-content-between mb-1.5">
            <span>
              <span className="ks-step-badge me-2">4</span>
              {t('village')}
            </span>
            {loadingVillage && (
              <span className="spinner-border spinner-border-sm text-primary" role="status" aria-hidden="true" />
            )}
          </label>
          <select
            id="village-select"
            className="form-select shadow-xs"
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
              <span>{errorVillage}</span>
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
