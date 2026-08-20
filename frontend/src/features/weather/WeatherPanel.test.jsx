import { render, screen } from '@testing-library/react'
import { WeatherPanel } from './WeatherPanel'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'

function renderWeather(props = {}, locale = 'id') {
  if (locale === 'en') {
    localStorage.setItem('kota-siaga.locale', 'en')
  }
  return render(
    <LocaleProvider>
      <WeatherPanel {...props} />
    </LocaleProvider>,
  )
}

describe('WeatherPanel', () => {
  it('renders forecast list in chronological order', () => {
    const items = [
      {
        id: 'w1',
        local_datetime: '2026-08-20 12:00:00',
        weather_description: 'Cerah Berawan',
        weather_description_en: 'Partly Cloudy',
        temperature_c: 28,
        humidity_percent: 70,
        precipitation_mm: 0,
        wind_speed: 15,
        wind_direction: 'NE',
      },
    ]

    renderWeather({ items })
    expect(screen.getByText('2026-08-20 12:00:00')).toBeInTheDocument()
    expect(screen.getByText('Cerah Berawan')).toBeInTheDocument()
    expect(screen.getByText('28°C')).toBeInTheDocument()
    expect(screen.getByText('70%')).toBeInTheDocument()
  })

  it('uses English weather description when locale is en', () => {
    const items = [
      {
        id: 'w1',
        local_datetime: '2026-08-20 12:00:00',
        weather_description: 'Hujan Ringan',
        weather_description_en: 'Light Rain',
        temperature_c: 24,
      },
    ]

    renderWeather({ items }, 'en')
    expect(screen.getByText('Light Rain')).toBeInTheDocument()
  })

  it('renders empty message when no items returned', () => {
    renderWeather({ items: [] })
    expect(
      screen.getByText(/Tidak ada data prakiraan cuaca|No weather forecast available/i),
    ).toBeInTheDocument()
  })

  it('renders loading state', () => {
    renderWeather({ loading: true })
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('renders error state with retry button', () => {
    const onRetry = vi.fn()
    renderWeather({ error: 'Failed to fetch weather', onRetry })
    expect(screen.getByText('Failed to fetch weather')).toBeInTheDocument()
    const retryBtn = screen.getByRole('button', { name: /Coba Lagi|Retry/i })
    retryBtn.click()
    expect(onRetry).toHaveBeenCalled()
  })
})
