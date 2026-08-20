import { listHospitals } from './hospitalService'
import * as client from '../../shared/api/client'

describe('hospitalService', () => {
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
})
