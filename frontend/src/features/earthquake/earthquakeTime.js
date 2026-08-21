function fallback(value) {
  return { date: value || '-', time: '' }
}

export function formatEarthquakeDateTime(value, locale = 'id') {
  const rawValue = String(value ?? '').trim()
  const match = rawValue.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})(?::\d{2}(?:\.\d+)?)?(Z|[+-]\d{2}:?\d{2})?$/)
  if (!match) return fallback(rawValue)

  const [, , , , , , timezone] = match
  const normalizedValue = rawValue.replace(' ', 'T')
  const date = new Date(`${normalizedValue}${timezone ? '' : '+07:00'}`)
  if (Number.isNaN(date.getTime())) return fallback(rawValue)

  const language = locale === 'en' ? 'en-US' : 'id-ID'

  return {
    date: new Intl.DateTimeFormat(language, {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      timeZone: 'Asia/Jakarta',
    }).format(date),
    time: new Intl.DateTimeFormat(language, {
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
      timeZone: 'Asia/Jakarta',
    }).format(date).replace(':', locale === 'en' ? ':' : '.'),
  }
}
