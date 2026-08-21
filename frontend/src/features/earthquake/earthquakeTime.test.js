import { formatEarthquakeDateTime } from './earthquakeTime'

describe('formatEarthquakeDateTime', () => {
  it('formats Indonesian event date and WIB time for non-technical readers', () => {
    expect(formatEarthquakeDateTime('2026-08-20 14:00:00', 'id')).toEqual({
      date: '20 Agustus 2026',
      time: '14.00',
    })
  })

  it('formats English event date while keeping the explicit WIB timezone', () => {
    expect(formatEarthquakeDateTime('2026-08-20T14:00:00', 'en')).toEqual({
      date: 'August 20, 2026',
      time: '14:00',
    })
  })

  it('converts BMKG UTC timestamps to WIB', () => {
    expect(formatEarthquakeDateTime('2026-08-21T01:18:05+00:00', 'id')).toEqual({
      date: '21 Agustus 2026',
      time: '08.18',
    })
  })

  it('returns the source value as a safe fallback when the timestamp is invalid', () => {
    expect(formatEarthquakeDateTime('not-a-timestamp', 'id')).toEqual({
      date: 'not-a-timestamp',
      time: '',
    })
  })
})
