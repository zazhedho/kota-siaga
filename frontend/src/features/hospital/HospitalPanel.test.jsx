import { act, fireEvent, render, screen } from '@testing-library/react'
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
  afterEach(() => {
    vi.useRealTimers()
    localStorage.removeItem('kota-siaga.locale')
  })

  it('renders an accessible search form and ignores one-character input', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()

    renderHospitals({ onSearch })

    expect(screen.getByRole('search', { name: /Cari rumah sakit|Search hospitals/i })).toBeInTheDocument()
    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    fireEvent.change(input, { target: { value: 'a' } })

    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    expect(onSearch).not.toHaveBeenCalled()
  })

  it('emits a trimmed search after 300ms when input has at least two characters', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()

    renderHospitals({ onSearch })

    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    fireEvent.change(input, { target: { value: '  RS  ' } })

    expect(onSearch).not.toHaveBeenCalled()
    await act(async () => {
      vi.advanceTimersByTime(299)
    })
    expect(onSearch).not.toHaveBeenCalled()

    await act(async () => {
      vi.advanceTimersByTime(1)
    })

    expect(onSearch).toHaveBeenCalledTimes(1)
    expect(onSearch).toHaveBeenCalledWith('RS')
  })

  it('does not emit one emoji but emits trimmed multi-code-point search', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()

    renderHospitals({ onSearch })

    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    fireEvent.change(input, { target: { value: '😀' } })

    await act(async () => {
      vi.advanceTimersByTime(300)
    })
    expect(onSearch).not.toHaveBeenCalled()

    fireEvent.change(input, { target: { value: '  😀a  ' } })
    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    expect(onSearch).toHaveBeenCalledTimes(1)
    expect(onSearch).toHaveBeenCalledWith('😀a')
  })

  it('caps controlled search input at 100 Unicode code points', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()
    const value = '😀'.repeat(101)

    renderHospitals({ onSearch })

    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    fireEvent.change(input, { target: { value } })

    expect([...input.value]).toHaveLength(100)
    expect(input).toHaveValue('😀'.repeat(100))

    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    expect(onSearch).toHaveBeenCalledWith('😀'.repeat(100))
  })

  it('submits a valid search immediately without a duplicate debounce call', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()

    renderHospitals({ onSearch })

    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    fireEvent.change(input, { target: { value: 'RS' } })
    fireEvent.submit(screen.getByRole('search', { name: /Cari rumah sakit|Search hospitals/i }))

    expect(onSearch).toHaveBeenCalledTimes(1)
    expect(onSearch).toHaveBeenCalledWith('RS')

    await act(async () => {
      vi.advanceTimersByTime(300)
    })
    expect(onSearch).toHaveBeenCalledTimes(1)
  })

  it('clears the search through the trailing clear action after the debounce', async () => {
    vi.useFakeTimers()
    const onSearch = vi.fn()

    renderHospitals({
      hospitals: [{ id: 'h1', name: 'RS' }],
      total: 1,
      search: 'RS',
      onSearch,
    })

    const input = screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })
    expect(input).toHaveValue('RS')
    fireEvent.click(screen.getByRole('button', { name: /Hapus pencarian|Clear search/i }))
    expect(input).toHaveValue('')

    await act(async () => {
      vi.advanceTimersByTime(300)
    })

    expect(onSearch).toHaveBeenCalledTimes(1)
    expect(onSearch).toHaveBeenCalledWith('')
  })

  it('shows a search-specific empty state without emergency links', () => {
    renderHospitals({ hospitals: [], total: 0, search: 'unknown', onSearch: vi.fn() })

    expect(screen.getByText(/Rumah sakit tidak ditemukan|No hospitals match your search/i)).toBeInTheDocument()
    expect(screen.getByText('Hapus pencarian', { selector: 'button' })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /119/ })).not.toBeInTheDocument()
  })

  it('announces displayed count without relying on unavailable total metadata', () => {
    renderHospitals({
      hospitals: [{ id: 'h1', name: 'RS 1' }, { id: 'h2', name: 'RS 2' }],
      total: 0,
    })

    const liveStatus = document.querySelector('#hospital-results .visually-hidden')
    expect(liveStatus).toHaveTextContent('Menampilkan 2 fasilitas kesehatan.')
    expect(liveStatus).not.toHaveTextContent('0')
  })

  it('announces search-empty state in the live status', () => {
    renderHospitals({ hospitals: [], total: 0, search: 'unknown', onSearch: vi.fn() })

    expect(document.querySelector('#hospital-results .visually-hidden')).toHaveTextContent(
      'Tidak ada rumah sakit yang cocok dengan pencarian ini.',
    )
  })

  it('announces errors in the live status', () => {
    renderHospitals({ error: 'Gagal memuat rumah sakit', onRetry: vi.fn() })

    expect(document.querySelector('#hospital-results .visually-hidden')).toHaveTextContent(
      'Gagal memuat rumah sakit',
    )
  })

  it('renders the English search label, placeholder, and hint', () => {
    localStorage.setItem('kota-siaga.locale', 'en')

    renderHospitals({ onSearch: vi.fn() })

    const input = screen.getByRole('textbox', { name: 'Search hospitals' })
    expect(input).toHaveAttribute('placeholder', 'Search by hospital name...')
    expect(screen.getByText('Type at least 2 characters to search hospitals.')).toBeInTheDocument()
  })

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

  it('keeps emergency contacts in one aligned action group', () => {
    renderHospitals({ hospitals: [], total: 0 })

    const emergencyActions = screen.getByRole('link', { name: /119/ }).parentElement

    expect(emergencyActions).toHaveClass('ks-emergency-actions')
    expect(screen.getByRole('link', { name: /119/ })).toHaveClass('ks-emergency-action')
    expect(screen.getByRole('link', { name: /112/ })).toHaveClass('ks-emergency-action')
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
