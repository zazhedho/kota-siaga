import { useEffect, useRef, useState } from 'react'
import { useLocale } from '../../shared/i18n'
import { AsyncLoading, AsyncError, AsyncEmpty } from '../../components/common/AsyncState'

const capSearch = (value) => (typeof value === 'string' ? Array.from(value).slice(0, 100).join('') : '')
const normalizeSearch = (value) => capSearch(value).trim()

export function HospitalPanel({
  hospitals = [],
  total = 0,
  loading = false,
  loadingMore = false,
  error = null,
  hasNextPage = false,
  search = '',
  onSearch,
  onLoadMore,
  onRetry,
}) {
  const { t } = useLocale()
  const normalizedSearch = normalizeSearch(search)
  const [inputValue, setInputValue] = useState(normalizedSearch)
  const searchTimerRef = useRef(null)
  const lastEmittedSearchRef = useRef(normalizedSearch)

  useEffect(() => {
    if (searchTimerRef.current !== null) {
      window.clearTimeout(searchTimerRef.current)
      searchTimerRef.current = null
    }

    const nextSearch = normalizeSearch(search)
    setInputValue(nextSearch)
    lastEmittedSearchRef.current = nextSearch
  }, [search])

  useEffect(() => () => {
    if (searchTimerRef.current !== null) window.clearTimeout(searchTimerRef.current)
  }, [])

  const emitSearch = (value) => {
    const nextSearch = normalizeSearch(value)
    if (nextSearch && Array.from(nextSearch).length < 2) return
    if (nextSearch === lastEmittedSearchRef.current) return

    lastEmittedSearchRef.current = nextSearch
    onSearch?.(nextSearch)
  }

  const scheduleSearch = (value) => {
    if (searchTimerRef.current !== null) window.clearTimeout(searchTimerRef.current)
    searchTimerRef.current = window.setTimeout(() => {
      searchTimerRef.current = null
      emitSearch(value)
    }, 300)
  }

  const handleSearchChange = (event) => {
    const nextValue = capSearch(event.target.value)
    setInputValue(nextValue)
    scheduleSearch(nextValue)
  }

  const handleSearchSubmit = (event) => {
    event.preventDefault()
    if (searchTimerRef.current !== null) {
      window.clearTimeout(searchTimerRef.current)
      searchTimerRef.current = null
    }
    emitSearch(inputValue)
  }

  const handleSearchClear = () => {
    setInputValue('')
    scheduleSearch('')
  }

  let liveStatus = t('hospitalResultsStatus', { count: hospitals.length })
  if (loading || loadingMore) {
    liveStatus = t('loading')
  } else if (error !== null && error !== undefined) {
    liveStatus = error ? `${t('genericError')}: ${error}` : t('genericError')
  } else if (hospitals.length === 0 && normalizedSearch) {
    liveStatus = t('hospitalSearchEmpty')
  }

  return (
    <div className="ks-card" aria-labelledby="hospitals-heading">
      <div className="ks-card-header">
        <div className="d-flex flex-wrap align-items-center justify-content-between gap-2">
          <div className="d-flex align-items-center gap-2">
            <div className="p-2 rounded-circle bg-primary-subtle text-primary d-inline-flex">
              <i className="bi bi-hospital fs-5" aria-hidden="true"></i>
            </div>
            <div>
              <h2 id="hospitals-heading" className="h6 mb-0 text-dark fw-bold">
                {t('hospitalTitle')}
              </h2>
              <span className="text-secondary small d-block">{t('hospitalSubtitle')}</span>
            </div>
          </div>
          {total > 0 && (
            <span className="badge bg-light text-dark border px-2 py-1 rounded-pill small">
              {t('showingHospitalsCount', { count: hospitals.length, total })}
            </span>
          )}
        </div>
      </div>

      <form
        className="ks-hospital-search"
        role="search"
        aria-label={t('hospitalSearchLabel')}
        onSubmit={handleSearchSubmit}
      >
        <label htmlFor="hospital-search-input" className="form-label fw-semibold small text-dark mb-1">
          {t('hospitalSearchLabel')}
        </label>
        <div className="ks-hospital-search-control">
          <i className="bi bi-search ks-hospital-search-icon" aria-hidden="true"></i>
          <input
            id="hospital-search-input"
            className="form-control"
            type="text"
            value={inputValue}
            placeholder={t('hospitalSearchPlaceholder')}
            aria-describedby="hospital-search-hint"
            onChange={handleSearchChange}
          />
          {inputValue && (
            <button
              type="button"
              className="ks-hospital-search-clear"
              aria-label={t('hospitalSearchClear')}
              onClick={handleSearchClear}
            >
              <i className="bi bi-x-lg" aria-hidden="true"></i>
            </button>
          )}
        </div>
        <p id="hospital-search-hint" className="small text-secondary mb-0 mt-1">
          {t('hospitalSearchHint')}
        </p>
      </form>

      <div id="hospital-results" aria-busy={loading || loadingMore}>
        <span className="visually-hidden" aria-live="polite">
          {liveStatus}
        </span>

      {loading && <AsyncLoading />}

      {!loading && error && <AsyncError message={error} onRetry={onRetry} />}

      {!loading && !error && hospitals.length === 0 && normalizedSearch && (
        <AsyncEmpty
          icon="bi-search"
          title={t('hospitalSearchEmptyTitle')}
          message={t('hospitalSearchEmpty')}
        >
          <button
            type="button"
            className="btn btn-link btn-sm text-decoration-none"
            onClick={handleSearchClear}
          >
            <i className="bi bi-x-circle me-1" aria-hidden="true"></i>
            {t('hospitalSearchClearAction')}
          </button>
        </AsyncEmpty>
      )}

      {!loading && !error && hospitals.length === 0 && !normalizedSearch && (
        <AsyncEmpty
          icon="bi-hospital"
          title={t('emptyHospitalsTitle')}
          message={t('emptyHospitals')}
        >
          <div className="ks-emergency-help p-3 bg-white border rounded shadow-xs mt-2 text-start mx-auto">
            <div className="d-flex align-items-center gap-2 mb-2">
              <i className="bi bi-telephone-fill text-danger"></i>
              <strong className="small text-dark">{t('emergencyHelpTitle')}</strong>
            </div>
            <div className="ks-emergency-actions">
              <a
                href="tel:119"
                className="ks-emergency-action btn btn-outline-danger btn-sm d-inline-flex align-items-center gap-2 rounded-pill px-3 py-1"
              >
                <i className="bi bi-telephone-outbound-fill"></i>
                <span>{t('emergencyCall119')}</span>
              </a>
              <a
                href="tel:112"
                className="ks-emergency-action btn btn-outline-secondary btn-sm d-inline-flex align-items-center gap-2 rounded-pill px-3 py-1"
              >
                <i className="bi bi-shield-shaded"></i>
                <span>{t('emergencyCall112')}</span>
              </a>
            </div>
          </div>
        </AsyncEmpty>
      )}

      {!loading && !error && hospitals.length > 0 && (
        <div>
          <div className={`d-flex flex-column gap-2 mb-3 ${hospitals.length > 4 ? 'ks-scrollable-list' : ''}`}>
            {hospitals.map((h, idx) => (
              <div key={h.id || idx} className="ks-hospital-card">
                <div className="d-flex flex-wrap align-items-start justify-content-between gap-2 mb-1.5">
                  <h3 className="h6 mb-0 fw-bold text-dark">{h.name}</h3>
                  <div className="d-flex gap-2 flex-wrap">
                    {h.type && (
                      <span className="badge bg-primary-subtle text-primary border border-primary-subtle rounded-pill small">
                        {h.type}
                      </span>
                    )}
                    {h.class && (
                      <span className="badge bg-secondary-subtle text-secondary-emphasis border border-secondary-subtle rounded-pill small">
                        {t('hospitalClass', { value: h.class })}
                      </span>
                    )}
                  </div>
                </div>

                {h.address && (
                  <p className="small text-secondary mb-2 lh-sm">
                    <i className="bi bi-geo-alt me-1 text-primary opacity-75" aria-hidden="true"></i>
                    {h.address}
                    {h.postal_code ? ` (${h.postal_code})` : ''}
                  </p>
                )}

                <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 pt-2 border-top border-light-subtle">
                  <div className="d-flex flex-wrap align-items-center gap-2 small">
                    {typeof h.beds_total === 'number' && h.beds_total > 0 && (
                      <span className="badge bg-light text-dark border d-inline-flex align-items-center gap-1">
                        <i className="bi bi-door-closed text-secondary" aria-hidden="true"></i>
                        {t('hospitalBeds', { count: h.beds_total })}
                      </span>
                    )}
                    {typeof h.icu_beds === 'number' && h.icu_beds > 0 && (
                      <span className="badge bg-danger-subtle text-danger border border-danger-subtle d-inline-flex align-items-center gap-1">
                        <i className="bi bi-heart-pulse-fill text-danger" aria-hidden="true"></i>
                        {t('hospitalICUBeds', { count: h.icu_beds })}
                      </span>
                    )}
                  </div>

                  {h.phone ? (
                    <a
                      href={`tel:${h.phone}`}
                      className="ks-hospital-phone-btn"
                      aria-label={t('callHospitalAria', { name: h.name })}
                    >
                      <i className="bi bi-telephone-fill" aria-hidden="true"></i>
                      <span>{h.phone}</span>
                    </a>
                  ) : (
                    <span className="small text-muted fst-italic">{t('noPhone')}</span>
                  )}
                </div>
              </div>
            ))}
          </div>

          {hasNextPage && (
            <div className="text-center pt-2">
              <button
                type="button"
                className="btn btn-outline-primary btn-sm px-4 rounded-pill shadow-xs"
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
    </div>
  )
}
