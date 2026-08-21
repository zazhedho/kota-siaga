import { render, screen } from '@testing-library/react'
import { EarthquakePanel } from './EarthquakePanel'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'

function renderEarthquakes(props = {}) {
  return render(
    <LocaleProvider>
      <EarthquakePanel {...props} />
    </LocaleProvider>,
  )
}

describe('EarthquakePanel', () => {
  it('renders earthquake item details', () => {
    const items = [
      {
        id: 'eq-1',
        date_time: '2026-08-20 14:00:00',
        magnitude: 5.4,
        depth_km: 15,
        latitude: -7.5,
        longitude: 107.2,
        region: 'Pusat gempa berada di laut 85 km Barat Daya Kab. Garut',
        potential: 'Tidak berpotensi tsunami',
        felt_areas: 'III Garut, II Bandung',
      },
    ]

    renderEarthquakes({ items })

    expect(screen.getByText('Waktu kejadian')).toBeInTheDocument()
    expect(screen.getByText('20 Agustus 2026')).toBeInTheDocument()
    expect(screen.getByText('14.00')).toBeInTheDocument()
    expect(screen.getByText('WIB (UTC+7)')).toBeInTheDocument()
    expect(screen.getByText('M 5.4')).toBeInTheDocument()
    expect(screen.getByText('15 km')).toBeInTheDocument()
    expect(screen.getByText(/Pusat gempa berada di laut/i)).toBeInTheDocument()
    expect(screen.getByText('Tidak berpotensi tsunami')).toBeInTheDocument()
    expect(screen.getByText('III Garut, II Bandung')).toBeInTheDocument()
  })

  it('renders empty message when no earthquakes', () => {
    renderEarthquakes({ items: [] })
    expect(
      screen.getByText(/Tidak ada data gempa bumi terkini|No recent earthquake data/i),
    ).toBeInTheDocument()
  })

  it('renders error and triggers retry', () => {
    const onRetry = vi.fn()
    renderEarthquakes({ error: 'Gagal memuat gempa', onRetry })

    expect(screen.getByText('Gagal memuat gempa')).toBeInTheDocument()
    const retryBtn = screen.getByRole('button', { name: /Coba Lagi|Retry/i })
    retryBtn.click()
    expect(onRetry).toHaveBeenCalled()
  })
})
