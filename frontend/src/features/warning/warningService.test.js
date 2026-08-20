import { listWarnings } from './warningService'
import * as client from '../../shared/api/client'

describe('warningService', () => {
  it('calls /warnings with provinsi query param', async () => {
    const mockData = [
      { id: '1', province: 'JAWA BARAT', event: 'Hujan Lebat Disertai Petir', severity: 'Moderate' },
    ]
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: mockData,
    })

    const result = await listWarnings('JAWA BARAT')
    expect(spy).toHaveBeenCalledWith('/warnings', {
      params: { provinsi: 'JAWA BARAT' },
      signal: undefined,
    })
    expect(result).toEqual(mockData)
  })
})
