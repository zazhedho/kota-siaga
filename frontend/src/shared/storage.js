const LOCALE_KEY = 'kota-siaga.locale'
const LOCATION_KEY = 'kota-siaga.location'

export function readLocale() {
  return localStorage.getItem(LOCALE_KEY) === 'en' ? 'en' : 'id'
}

export function writeLocale(locale) {
  localStorage.setItem(LOCALE_KEY, locale === 'en' ? 'en' : 'id')
}

export function readStoredLocation() {
  try {
    const value = JSON.parse(localStorage.getItem(LOCATION_KEY) || 'null')
    return value && value.provinceId && value.cityId && value.districtId && value.villageId && value.adm4
      ? value
      : null
  } catch {
    clearStoredLocation()
    return null
  }
}

export function writeStoredLocation(location) {
  localStorage.setItem(
    LOCATION_KEY,
    JSON.stringify({
      provinceId: location.province.id,
      cityId: location.city.id,
      districtId: location.district.id,
      villageId: location.village.id,
      adm4: location.village.code,
    }),
  )
}

export function clearStoredLocation() {
  localStorage.removeItem(LOCATION_KEY)
}
