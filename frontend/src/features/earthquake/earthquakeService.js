import { request } from '../../shared/api/client'
import { readData } from '../../shared/api/response'

export async function listLatest(signal) {
  const body = await request('/earthquakes/latest', { signal })
  const data = readData(body)
  if (Array.isArray(data)) return data
  if (data && typeof data === 'object') return [data]
  return []
}
