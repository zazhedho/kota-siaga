import { request } from '../../shared/api/client'
import { readData } from '../../shared/api/response'

function forecastTimestamp(item) {
  const timestamp = Date.parse(item.local_datetime || item.datetime || '')
  return Number.isNaN(timestamp) ? -Infinity : timestamp
}

export async function getForecast(adm4, signal) {
  const body = await request('/weather', { params: { adm4 }, signal })
  const data = readData(body)
  if (!Array.isArray(data)) return []
  return [...data].sort((left, right) => forecastTimestamp(right) - forecastTimestamp(left))
}
