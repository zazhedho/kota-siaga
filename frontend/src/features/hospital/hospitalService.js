import { request } from '../../shared/api/client'
import { readPage } from '../../shared/api/response'

export function listHospitals(kabupatenId, page = 1, search = '', signal) {
  if (typeof search !== 'string') {
    signal = search
    search = ''
  }

  const normalizedSearch = search.trim()
  return request('/hospitals', {
    params: {
      kabupaten_id: kabupatenId,
      page,
      per_page: 20,
      ...(normalizedSearch ? { search: normalizedSearch } : {}),
    },
    signal,
  }).then(readPage)
}
