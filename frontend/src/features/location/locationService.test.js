import {
	listProvinces,
	listCities,
	listDistricts,
	listVillages,
	searchLocations,
	resolveLocation,
} from './locationService'
import * as client from '../../shared/api/client'

describe('locationService', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('calls listProvinces with page=1 and per_page=100', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [{ id: '32', name: 'JAWA BARAT' }],
      total_data: 34,
      total_pages: 1,
      current_page: 1,
      limit: 100,
    })

    const result = await listProvinces()
    expect(spy).toHaveBeenCalledWith('/locations/province', {
      params: { page: 1, per_page: 100 },
      signal: undefined,
    })
    expect(result.rows).toHaveLength(1)
    expect(result.rows[0].name).toBe('JAWA BARAT')
  })

  it('calls listCities with province_id', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [{ id: '3273', name: 'KOTA BANDUNG' }],
    })

    const result = await listCities('32')
    expect(spy).toHaveBeenCalledWith('/locations/city', {
      params: { province_id: '32', page: 1, per_page: 100 },
      signal: undefined,
    })
    expect(result.rows[0].name).toBe('KOTA BANDUNG')
  })

  it('calls listDistricts with kabupaten_id', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [{ id: '3273010', name: 'SUKAJADI' }],
    })

    const result = await listDistricts('3273')
    expect(spy).toHaveBeenCalledWith('/locations/district', {
      params: { kabupaten_id: '3273', page: 1, per_page: 100 },
      signal: undefined,
    })
    expect(result.rows[0].name).toBe('SUKAJADI')
  })

	it('calls listVillages with kecamatan_id', async () => {
    const spy = vi.spyOn(client, 'request').mockResolvedValue({
      status: true,
      data: [{ id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' }],
    })

    const result = await listVillages('3273010')
    expect(spy).toHaveBeenCalledWith('/locations/village', {
      params: { kecamatan_id: '3273010', page: 1, per_page: 100 },
      signal: undefined,
    })
    expect(result.rows[0].name).toBe('PASTEUR')
		expect(result.rows[0].code).toBe('32.73.01.1001')
	})

	it('searches locations with a bounded query', async () => {
		const signal = new AbortController().signal
		const spy = vi.spyOn(client, 'request').mockResolvedValue({
			status: true,
			data: [{ id: '3273011001', code: '32.73.01.1001', name: 'PASTEUR', level: 'village' }],
		})

		const result = await searchLocations('pasteur', 10, signal)
		expect(spy).toHaveBeenCalledWith('/locations/search', {
			params: { q: 'pasteur', limit: 10 },
			signal,
		})
		expect(result[0].name).toBe('PASTEUR')
	})

	it('resolves a selected village path', async () => {
		const signal = new AbortController().signal
		const spy = vi.spyOn(client, 'request').mockResolvedValue({
			status: true,
			data: { village: { id: '3273011001', name: 'PASTEUR' } },
		})

		const result = await resolveLocation('32.73.01.1001', signal)
		expect(spy).toHaveBeenCalledWith('/locations/resolve', {
			params: { code: '32.73.01.1001' },
			signal,
		})
		expect(result.village.name).toBe('PASTEUR')
	})
})
