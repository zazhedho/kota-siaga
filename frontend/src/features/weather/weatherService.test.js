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

  it('returns forecasts in descending local time order', async () => {
    const mockData = [
      { id: 'morning', local_datetime: '2026-08-20T06:00:00+07:00' },
      { id: 'latest', local_datetime: '2026-08-20T18:00:00+07:00' },
      { id: 'afternoon', local_datetime: '2026-08-20T14:00:00+07:00' },
    ]
    vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: mockData,
    })

    const result = await getForecast('32.73.01.1001')

    expect(result.map((item) => item.id)).toEqual(['latest', 'afternoon', 'morning'])
  })
})
