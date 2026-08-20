import { request } from '../../shared/api/client'
import { readData } from '../../shared/api/response'

export async function listWarnings(provinceName, signal) {
  const body = await request('/warnings', { params: { provinsi: provinceName }, signal })
  return readData(body) ?? []
}
