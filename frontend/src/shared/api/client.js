const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/$/, '')

export class ApiError extends Error {
  constructor(message, { status = 0, code = 0 } = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function getApiErrorMessage(error, t) {
  const translate = typeof t === 'function' ? t : (key) => key

  if (!error) return ''
  if (typeof error === 'string') return error
  if (error?.name === 'AbortError') return ''
  if (error instanceof ApiError || typeof error?.status === 'number') {
    if (error.status === 429) return translate('rateLimitError')
    if (error.status >= 500 || error.status === 0) return translate('serviceUnavailable')
    return translate('genericError')
  }

  return error?.message || translate('genericError')
}

export async function request(path, { params, signal } = {}) {
  const url = new URL(`${API_BASE_URL}${path}`, window.location.origin)
  Object.entries(params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      url.searchParams.set(key, value)
    }
  })

  let response
  try {
    response = await fetch(url, {
      signal,
      headers: { Accept: 'application/json' },
    })
  } catch (error) {
    if (error.name === 'AbortError') throw error
    throw new ApiError('Network request failed')
  }

  let body
  try {
    body = await response.json()
  } catch {
    throw new ApiError('Request returned invalid JSON', { status: response.status })
  }

  if (!response.ok || body?.status === false) {
    throw new ApiError(
      body?.error?.message || body?.message || 'Request failed',
      {
        status: response.status,
        code: body?.error?.code || body?.code || 0,
      },
    )
  }

  return body
}
