import { getForecast } from './weatherService'
import * as client from '../../shared/api/client'

describe('weatherService', () => {
  it('calls /weather with adm4 parameter', async () => {
    const mockData = [
      { id: '1', adm4: '32.73.01.1001', temperature_c: 26, weather: 'Cerah Berawan' },
    ]
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: mockData,
    })

    const result = await getForecast('32.73.01.1001')
    expect(spy).toHaveBeenCalledWith('/weather', {
      params: { adm4: '32.73.01.1001' },
      signal: undefined,
    })
    expect(result).toEqual(mockData)
  })
})
