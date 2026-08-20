import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HospitalPanel } from './HospitalPanel'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'

function renderHospitals(props = {}) {
  return render(
    <LocaleProvider>
      <HospitalPanel {...props} />
    </LocaleProvider>,
  )
}

describe('HospitalPanel', () => {
  it('renders hospital cards with details', () => {
    const hospitals = [
      {
        id: 'h1',
        name: 'RSUP Dr. Hasan Sadikin',
        type: 'RSUP',
        class: 'A',
        address: 'Jl. Pasteur No. 38',
        phone: '022-2034953',
        beds_total: 800,
        icu_beds: 50,
      },
    ]

    renderHospitals({ hospitals, total: 1 })

    expect(screen.getByText('RSUP Dr. Hasan Sadikin')).toBeInTheDocument()
    expect(screen.getByText('RSUP')).toBeInTheDocument()
    expect(screen.getByText('Kelas A')).toBeInTheDocument()
    expect(screen.getByText(/Jl. Pasteur No. 38/)).toBeInTheDocument()
    expect(screen.getByText('022-2034953')).toBeInTheDocument()
    expect(screen.getByText(/800 tempat tidur/)).toBeInTheDocument()
    expect(screen.getByText(/50 ICU/)).toBeInTheDocument()
  })

  it('renders load more button when hasNextPage is true', async () => {
    const onLoadMore = vi.fn()
    const user = userEvent.setup()

    renderHospitals({
      hospitals: [{ id: 'h1', name: 'RS 1' }],
      total: 30,
      hasNextPage: true,
      onLoadMore,
    })

    const loadMoreBtn = screen.getByRole('button', { name: /Muat Lebih Banyak|Load More/i })
    expect(loadMoreBtn).toBeInTheDocument()

    await user.click(loadMoreBtn)
    expect(onLoadMore).toHaveBeenCalled()
  })

  it('renders empty message when no hospitals found', () => {
    renderHospitals({ hospitals: [], total: 0 })
    expect(
      screen.getByText(/Tidak ada data fasilitas kesehatan|No healthcare facilities found/i),
    ).toBeInTheDocument()
  })

  it('renders error state with retry button', () => {
    const onRetry = vi.fn()
    renderHospitals({ error: 'Gagal memuat rumah sakit', onRetry })

    expect(screen.getByText('Gagal memuat rumah sakit')).toBeInTheDocument()
    const retryBtn = screen.getByRole('button', { name: /Coba Lagi|Retry/i })
    retryBtn.click()
    expect(onRetry).toHaveBeenCalled()
  })

  it('translates hospital detail labels in English', () => {
    localStorage.setItem('kota-siaga.locale', 'en')

    renderHospitals({
      hospitals: [{ id: 'h1', name: 'RS 1', class: 'A', beds_total: 10, icu_beds: 2 }],
      total: 1,
    })

    expect(screen.getByText('Class A')).toBeInTheDocument()
    expect(screen.getByText('10 beds')).toBeInTheDocument()
    expect(screen.getByText('2 ICU beds')).toBeInTheDocument()
    expect(screen.queryByText('Kelas A')).not.toBeInTheDocument()
  })
})
