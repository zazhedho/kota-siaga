import { request } from '../../shared/api/client'
import { readPage } from '../../shared/api/response'

export function listHospitals(kabupatenId, page = 1, signal) {
  return request('/hospitals', {
    params: { kabupaten_id: kabupatenId, page, per_page: 20 },
    signal,
  }).then(readPage)
}
