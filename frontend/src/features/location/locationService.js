import { request } from '../../shared/api/client'
import { readData, readPage } from '../../shared/api/response'

async function list(path, params, signal) {
  const body = await request(path, {
    params: { ...params, page: 1, per_page: 100 },
    signal,
  })
  return readPage(body)
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
  return list('/locations/village', { kecamatan_id: districtId }, signal)
}

export async function searchLocations(query, limit = 10, signal) {
  const body = await request('/locations/search', {
    params: { q: query, limit },
    signal,
  })
  return readData(body) || []
}

export async function resolveLocation(code, signal) {
  const body = await request('/locations/resolve', {
    params: { code },
    signal,
  })
  return readData(body)
}
