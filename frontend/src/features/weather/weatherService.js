import { request } from '../../shared/api/client'
import { readData } from '../../shared/api/response'

export async function getForecast(adm4, signal) {
  const body = await request('/weather', { params: { adm4 }, signal })
  return readData(body) ?? []
}
