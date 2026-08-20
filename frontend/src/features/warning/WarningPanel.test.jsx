import { render, screen } from '@testing-library/react'
import { WarningPanel } from './WarningPanel'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'

function renderWarnings(props = {}) {
  return render(
    <LocaleProvider>
      <WarningPanel {...props} />
    </LocaleProvider>,
  )
}

describe('WarningPanel', () => {
  it('renders active warning cards with event and severity', () => {
    const warnings = [
      {
        id: 'warn-1',
        event: 'Peringatan Hujan Lebat',
        headline: 'Waspada Banjir dan Longsor',
        severity: 'Severe',
        description: 'Potensi hujan lebat disertai kilat/petir di wilayah pegunungan.',
        instruction: 'Masyarakat dihimbau menjauhi tebing curam.',
        effective: '2026-08-20 10:00',
        expires: '2026-08-20 18:00',
        source: 'BMKG',
      },
    ]

    renderWarnings({ warnings })

    expect(screen.getByText('Peringatan Hujan Lebat')).toBeInTheDocument()
    expect(screen.getByText('Waspada Banjir dan Longsor')).toBeInTheDocument()
    expect(screen.getByText(/Awas \(Berat\)|Severe Warning/i)).toBeInTheDocument()
    expect(screen.getByText(/Masyarakat dihimbau menjauhi tebing curam/i)).toBeInTheDocument()
  })

  it('renders empty calm state when no warnings exist', () => {
    renderWarnings({ warnings: [] })
    expect(
      screen.getByText(/Tidak ada peringatan dini cuaca aktif|No active weather warnings/i),
    ).toBeInTheDocument()
  })

  it('renders error and triggers retry', () => {
    const onRetry = vi.fn()
    renderWarnings({ error: 'Gagal memuat peringatan', onRetry })

    expect(screen.getByText('Gagal memuat peringatan')).toBeInTheDocument()
    const retryBtn = screen.getByRole('button', { name: /Coba Lagi|Retry/i })
    retryBtn.click()
    expect(onRetry).toHaveBeenCalled()
  })

  it('translates warning UI labels in English without translating source content', async () => {
    localStorage.setItem('kota-siaga.locale', 'en')

    renderWarnings({
      warnings: [{ id: 'warn-1', instruction: 'Ikuti arahan petugas.' }],
    })

    expect(screen.getByText('Instruction:')).toBeInTheDocument()
    expect(screen.getByText('Weather Warning')).toBeInTheDocument()
    expect(screen.getByText('Ikuti arahan petugas.')).toBeInTheDocument()
  })
})
