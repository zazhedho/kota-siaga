import { request, ApiError, getApiErrorMessage } from './client'

describe('request client', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('unwraps successful json response', async () => {
    const mockData = { code: 200, status: true, message: 'Success', data: [{ id: '1' }] }
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => mockData,
    })

    const result = await request('/locations/province')
    expect(result).toEqual(mockData)
    expect(global.fetch).toHaveBeenCalledWith(
      expect.objectContaining({
        href: expect.stringContaining('/api/locations/province'),
      }),
      expect.objectContaining({
        headers: { Accept: 'application/json' },
      }),
    )
  })

  it('encodes query params properly', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: true, data: [] }),
    })

    await request('/locations/city', { params: { province_id: '32', page: 1, empty: null } })
    expect(global.fetch).toHaveBeenCalledWith(
      expect.objectContaining({
        href: expect.stringContaining('/api/locations/city?province_id=32&page=1'),
      }),
      expect.any(Object),
    )
  })

  it('throws ApiError when response is not ok', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({
        code: 404,
        status: false,
        error: { code: 404, message: 'Not found' },
      }),
    })

    await expect(request('/unknown')).rejects.toThrow('Not found')
    await expect(request('/unknown')).rejects.toBeInstanceOf(ApiError)
  })

  it('throws ApiError on non-JSON response', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 502,
      json: async () => {
        throw new Error('Unexpected token < in JSON')
      },
    })

    await expect(request('/broken')).rejects.toThrow('Request returned invalid JSON')
  })

  it('propagates AbortError when signal aborted', async () => {
    const abortError = new Error('Aborted')
    abortError.name = 'AbortError'
    global.fetch = vi.fn().mockRejectedValue(abortError)

    const controller = new AbortController()
    controller.abort()

    await expect(request('/aborted', { signal: controller.signal })).rejects.toThrow('Aborted')
  })

  it('maps API failures to translated safe messages', () => {
    const t = vi.fn((key) => key)

    expect(getApiErrorMessage(new ApiError('upstream detail', { status: 502 }), t)).toBe('serviceUnavailable')
    expect(getApiErrorMessage(new ApiError('quota detail', { status: 429 }), t)).toBe('rateLimitError')
    expect(t).toHaveBeenCalledWith('serviceUnavailable')
    expect(t).toHaveBeenCalledWith('rateLimitError')
  })

  it('preserves local non-API error messages', () => {
    const t = vi.fn((key) => key)

    expect(getApiErrorMessage(new Error('local failure'), t)).toBe('local failure')
  })
})
