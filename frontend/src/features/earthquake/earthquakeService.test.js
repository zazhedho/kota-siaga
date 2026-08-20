import { listLatest } from './earthquakeService'
import * as client from '../../shared/api/client'

describe('earthquakeService', () => {
  it('calls /earthquakes/latest and returns array', async () => {
    const mockData = [
      { id: 'eq1', magnitude: 5.2, depth_km: 10, region: 'Selatan Jawa' },
    ]
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: mockData,
    })

    const result = await listLatest()
    expect(spy).toHaveBeenCalledWith('/earthquakes/latest', { signal: undefined })
    expect(result).toEqual(mockData)
  })

  it('handles single object response as array', async () => {
    const mockItem = { id: 'eq1', magnitude: 5.2 }
    vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: mockItem,
    })

    const result = await listLatest()
    expect(result).toEqual([mockItem])
  })
})
