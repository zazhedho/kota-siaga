import { useState, useEffect, useRef, useCallback } from 'react'
import { useLocale } from '../../shared/i18n'
import { ComboBox } from '../../components/common/ComboBox'
import { getApiErrorMessage } from '../../shared/api/client'
import {
	listProvinces,
	listCities,
	listDistricts,
	listVillages,
	searchLocations,
	resolveLocation,
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

  const [directSearchOpen, setDirectSearchOpen] = useState(false)
  const [directSearchQuery, setDirectSearchQuery] = useState('')
  const [directSearchOptions, setDirectSearchOptions] = useState([])
  const [directSearchValue, setDirectSearchValue] = useState('')
  const [loadingDirectSearch, setLoadingDirectSearch] = useState(false)
  const [errorDirectSearch, setErrorDirectSearch] = useState(null)

  const provinceAbortRef = useRef(null)
  const cityAbortRef = useRef(null)
  const districtAbortRef = useRef(null)
  const villageAbortRef = useRef(null)
  const directSearchTimerRef = useRef(null)
  const directSearchAbortRef = useRef(null)
  const directResolveAbortRef = useRef(null)
  const directSearchRequestRef = useRef(0)
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
      if (directSearchTimerRef.current) window.clearTimeout(directSearchTimerRef.current)
      if (directSearchAbortRef.current) directSearchAbortRef.current.abort()
      if (directResolveAbortRef.current) directResolveAbortRef.current.abort()
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

  const clearDirectSearch = useCallback(() => {
    if (directSearchTimerRef.current) window.clearTimeout(directSearchTimerRef.current)
    if (directSearchAbortRef.current) directSearchAbortRef.current.abort()
    if (directResolveAbortRef.current) directResolveAbortRef.current.abort()
    directSearchRequestRef.current += 1
    setDirectSearchOpen(false)
    setDirectSearchQuery('')
    setDirectSearchOptions([])
    setDirectSearchValue('')
    setLoadingDirectSearch(false)
    setErrorDirectSearch(null)
  }, [])

  const applyLocationPath = useCallback((path) => {
    if (!path?.province || !path?.city || !path?.district) return false

    setProvinces((current) => (
      current.some((item) => String(item.id) === String(path.province.id))
        ? current
        : [...current, path.province]
    ))
    setCities([path.city])
    setDistricts([path.district])
    setSelectedProvinceId(String(path.province.id))
    setSelectedCityId(String(path.city.id))
    setSelectedDistrictId(String(path.district.id))

    if (path.village) {
      setVillages([path.village])
      setSelectedVillageId(String(path.village.id))
      onComplete?.({ ...path, adm4: path.village.code })
    } else {
      setVillages([])
      setSelectedVillageId('')
      onComplete?.(null)
      fetchVillages(path.district.id)
    }
    return true
  }, [fetchVillages, onComplete])

  const handleDirectSearchQuery = useCallback((query) => {
    setDirectSearchQuery(query)
    setErrorDirectSearch(null)
    if (directSearchTimerRef.current) window.clearTimeout(directSearchTimerRef.current)
    if (directSearchAbortRef.current) directSearchAbortRef.current.abort()
    directSearchRequestRef.current += 1
    const requestId = directSearchRequestRef.current
    const normalizedQuery = query.trim()
    setDirectSearchOptions([])

    if (normalizedQuery.length < 3) {
      setLoadingDirectSearch(false)
      return
    }

    setLoadingDirectSearch(true)
    directSearchTimerRef.current = window.setTimeout(async () => {
      const controller = new AbortController()
      directSearchAbortRef.current = controller

      try {
        const results = await searchLocations(normalizedQuery, 10, controller.signal)
        if (controller.signal.aborted || requestId !== directSearchRequestRef.current) return
        setDirectSearchOptions(results.map((item) => ({
          id: item.code || item.id,
          name: item.hierarchy || item.name,
        })))
      } catch (err) {
        if (!controller.signal.aborted && requestId === directSearchRequestRef.current) {
          setErrorDirectSearch(getApiErrorMessage(err, t) || t('directSearchError'))
        }
      } finally {
        if (requestId === directSearchRequestRef.current) setLoadingDirectSearch(false)
      }
    }, 300)
  }, [t])

  const handleDirectSearchSelect = useCallback(async (code) => {
    if (!code) return
    if (directSearchTimerRef.current) window.clearTimeout(directSearchTimerRef.current)
    if (directSearchAbortRef.current) directSearchAbortRef.current.abort()
    if (directResolveAbortRef.current) directResolveAbortRef.current.abort()

    const controller = new AbortController()
    directResolveAbortRef.current = controller
    setDirectSearchValue(code)
    setLoadingDirectSearch(true)
    setErrorDirectSearch(null)

    try {
      const path = await resolveLocation(code, controller.signal)
      if (controller.signal.aborted) return
      if (!applyLocationPath(path)) {
        setErrorDirectSearch(t('directSearchError'))
        return
      }
      setDirectSearchOpen(false)
      setDirectSearchQuery('')
      setDirectSearchOptions([])
      setDirectSearchValue('')
    } catch (err) {
      if (!controller.signal.aborted) {
        setErrorDirectSearch(getApiErrorMessage(err, t) || t('directSearchError'))
      }
    } finally {
      if (!controller.signal.aborted) setLoadingDirectSearch(false)
    }
  }, [applyLocationPath, t])

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

  const handleProvinceChange = (pId) => {
    clearDirectSearch()
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

  const handleCityChange = (cId) => {
    clearDirectSearch()
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

  const handleDistrictChange = (dId) => {
    clearDirectSearch()
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

  const handleVillageChange = (vId) => {
    clearDirectSearch()
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
        <button
          type="button"
          className="btn btn-sm btn-outline-primary ks-location-search-toggle d-inline-flex align-items-center gap-2"
          aria-controls="direct-location-search"
          aria-expanded={directSearchOpen}
          aria-pressed={directSearchOpen}
          onClick={() => (directSearchOpen ? clearDirectSearch() : setDirectSearchOpen(true))}
        >
          <i className="bi bi-search" aria-hidden="true"></i>
          <span>{t('directSearchToggle')}</span>
        </button>
      </div>

      {directSearchOpen && (
        <div id="direct-location-search" className="ks-direct-search mb-3">
          <label htmlFor="location-search" className="form-label fw-semibold small text-dark mb-1">
            {t('directSearchLabel')}
          </label>
          <ComboBox
            id="location-search"
            value={directSearchValue}
            options={directSearchOptions}
            placeholder={t('directSearchPlaceholder')}
            noResultsLabel={directSearchQuery.trim().length < 3 ? t('directSearchHint') : t('directSearchNoResults')}
            loadingLabel={t('directSearchLoading')}
            ariaBusy={loadingDirectSearch}
            onQueryChange={handleDirectSearchQuery}
            onChange={handleDirectSearchSelect}
          />
          <p className="small text-secondary mb-0 mt-1">{t('directSearchHint')}</p>
          {errorDirectSearch && (
            <div className="small text-danger mt-1" role="alert">
              {errorDirectSearch}
            </div>
          )}
        </div>
      )}

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
          <ComboBox
            id="province-select"
            value={selectedProvinceId}
            options={provinces}
            placeholder={loadingProvince ? t('loadingLocations') : t('selectProvince')}
            disabled={loadingProvince}
            noResultsLabel={t('noLocationResults')}
            onChange={handleProvinceChange}
            ariaBusy={loadingProvince}
          />
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
          <ComboBox
            id="city-select"
            value={selectedCityId}
            options={cities}
            placeholder={loadingCity ? t('loadingLocations') : t('selectCity')}
            disabled={!selectedProvinceId || loadingCity}
            noResultsLabel={t('noLocationResults')}
            onChange={handleCityChange}
            ariaBusy={loadingCity}
          />
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
          <ComboBox
            id="district-select"
            value={selectedDistrictId}
            options={districts}
            placeholder={loadingDistrict ? t('loadingLocations') : t('selectDistrict')}
            disabled={!selectedCityId || loadingDistrict}
            noResultsLabel={t('noLocationResults')}
            onChange={handleDistrictChange}
            ariaBusy={loadingDistrict}
          />
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
          <ComboBox
            id="village-select"
            value={selectedVillageId}
            options={villages}
            placeholder={loadingVillage ? t('loadingLocations') : t('selectVillage')}
            disabled={!selectedDistrictId || loadingVillage}
            noResultsLabel={t('noLocationResults')}
            onChange={handleVillageChange}
            ariaBusy={loadingVillage}
          />
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
