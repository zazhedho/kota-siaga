import { listHospitals } from './hospitalService'
import * as client from '../../shared/api/client'

describe('hospitalService', () => {
  afterEach(() => vi.restoreAllMocks())

  it('calls /hospitals with kabupaten_id, page and per_page=20', async () => {
    const mockResponse = {
      status: true,
      data: [{ id: 'h1', name: 'RSUD Hasan Sadikin' }],
      total_data: 25,
      total_pages: 2,
      current_page: 1,
      limit: 20,
      next_page: true,
      prev_page: false,
    }
    const spy = vi.spyOn(client, 'request').mockResolvedValue(mockResponse)

    const result = await listHospitals('3273', 1)
    expect(spy).toHaveBeenCalledWith('/hospitals', {
      params: { kabupaten_id: '3273', page: 1, per_page: 20 },
      signal: undefined,
    })
    expect(result.rows).toHaveLength(1)
    expect(result.nextPage).toBe(true)
  })

  it('sends a trimmed search and omits the search key when empty', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [],
      total_pages: 1,
      current_page: 1,
    })
    const signal = new AbortController().signal

    await listHospitals('3174', 1, '  Mayapada  ', signal)
    expect(spy).toHaveBeenLastCalledWith('/hospitals', {
      params: { kabupaten_id: '3174', page: 1, per_page: 20, search: 'Mayapada' },
      signal,
    })

    await listHospitals('3174', 2, '   ', signal)
    expect(spy).toHaveBeenLastCalledWith('/hospitals', {
      params: { kabupaten_id: '3174', page: 2, per_page: 20 },
      signal,
    })
  })

  it('preserves a legacy third-argument signal when search is omitted', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [],
      total_pages: 1,
      current_page: 1,
    })
    const signal = new AbortController().signal

    await listHospitals('3273', 1, signal)

    expect(spy).toHaveBeenCalledWith('/hospitals', {
      params: { kabupaten_id: '3273', page: 1, per_page: 20 },
      signal,
    })
  })
})