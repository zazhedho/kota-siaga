import { request } from '../../shared/api/client'
import { readPage } from '../../shared/api/response'

async function list(path, params, signal) {
  const body = await request(path, {
    params: { ...params, page: 1, per_page: 100 },
    signal,
  })
  return readPage(body)
}

export function formatAdm4(code) {
  const value = String(code ?? '')
  return /^\d{10}$/.test(value)
    ? `${value.slice(0, 2)}.${value.slice(2, 4)}.${value.slice(4, 6)}.${value.slice(6)}`
    : value
}

export function listProvinces(signal) {
  return list('/locations/province', {}, signal)
}

export function listCities(provinceId, signal) {
  return list('/locations/city', { province_id: provinceId }, signal)
}

export function listDistricts(cityId, signal) {
  return list('/locations/district', { kabupaten_id: cityId }, signal)
}

export function listVillages(districtId, signal) {
  return list('/locations/village', { kecamatan_id: districtId }, signal).then((result) => ({
    ...result,
    rows: result.rows.map((village) => ({
      ...village,
      code: formatAdm4(village.code),
    })),
  }))
}
