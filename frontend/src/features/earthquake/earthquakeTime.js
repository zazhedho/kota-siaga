function fallback(value) {
  return { date: value || '-', time: '' }
}

export function formatEarthquakeDateTime(value, locale = 'id') {
  const rawValue = String(value ?? '').trim()
  const match = rawValue.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})/)
  if (!match) return fallback(rawValue)

  const [, yearValue, monthValue, dayValue, hour, minute] = match
  const year = Number(yearValue)
  const month = Number(monthValue)
  const day = Number(dayValue)
  const date = new Date(Date.UTC(year, month - 1, day))

  if (
    Number.isNaN(date.getTime()) ||
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  ) {
    return fallback(rawValue)
  }

  return {
    date: new Intl.DateTimeFormat(locale === 'en' ? 'en-US' : 'id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(date),
    time: locale === 'en' ? `${hour}:${minute}` : `${hour}.${minute}`,
  }
}
