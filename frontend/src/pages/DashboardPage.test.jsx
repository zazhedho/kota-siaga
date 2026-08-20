import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DashboardPage } from './DashboardPage'
import { LocaleProvider } from '../shared/i18n/LocaleProvider'
import * as locationService from '../features/location/locationService'
import * as weatherService from '../features/weather/weatherService'
import * as warningService from '../features/warning/warningService'
import * as earthquakeService from '../features/earthquake/earthquakeService'
import * as hospitalService from '../features/hospital/hospitalService'

vi.mock('../features/location/locationService', () => ({
  listProvinces: vi.fn(),
  listCities: vi.fn(),
  listDistricts: vi.fn(),
  listVillages: vi.fn(),
}))

vi.mock('../features/weather/weatherService', () => ({
  getForecast: vi.fn(),
}))

vi.mock('../features/warning/warningService', () => ({
  listWarnings: vi.fn(),
}))

vi.mock('../features/earthquake/earthquakeService', () => ({
  listLatest: vi.fn(),
}))

vi.mock('../features/hospital/hospitalService', () => ({
  listHospitals: vi.fn(),
}))

function renderDashboard() {
  return render(
    <LocaleProvider>
      <DashboardPage />
    </LocaleProvider>,
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    localStorage.clear()

    locationService.listProvinces.mockResolvedValue({
      rows: [{ id: '32', name: 'JAWA BARAT' }],
    })
    locationService.listCities.mockResolvedValue({
      rows: [{ id: '3273', name: 'KOTA BANDUNG' }],
    })
    locationService.listDistricts.mockResolvedValue({
      rows: [{ id: '3273010', name: 'SUKAJADI' }],
    })
    locationService.listVillages.mockResolvedValue({
      rows: [{ id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' }],
    })

    weatherService.getForecast.mockResolvedValue([
      {
        id: 'w1',
        local_datetime: '2026-08-20 12:00:00',
        weather_description: 'Cerah',
        temperature_c: 27,
      },
    ])

    warningService.listWarnings.mockResolvedValue([
      {
        id: 'warn1',
        event: 'Peringatan Hujan Lebat',
        severity: 'Moderate',
      },
    ])

    earthquakeService.listLatest.mockResolvedValue([
      {
        id: 'eq1',
        magnitude: 5.0,
        region: 'Selatan Jawa Barat',
      },
    ])

    hospitalService.listHospitals.mockResolvedValue({
      rows: [{ id: 'h1', name: 'RS Hasan Sadikin' }],
      total: 1,
      page: 1,
      nextPage: false,
    })
  })

  it('renders initial prompt and starts all requests upon village selection', async () => {
    const user = userEvent.setup()
    renderDashboard()

    expect(
      screen.getByText(/Silakan pilih lokasi lengkap|Please select a complete location/i),
    ).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Provinsi|Province/i), '32')
    await waitFor(() => expect(screen.getByRole('option', { name: 'KOTA BANDUNG' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kabupaten\/Kota|City or regency/i), '3273')
    await waitFor(() => expect(screen.getByRole('option', { name: 'SUKAJADI' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kecamatan|District/i), '3273010')
    await waitFor(() => expect(screen.getByRole('option', { name: 'PASTEUR' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kelurahan\/Desa|Village/i), '3273010100')

    await waitFor(() => {
      expect(weatherService.getForecast).toHaveBeenCalledWith('32.73.01.1001', expect.any(Object))
      expect(warningService.listWarnings).toHaveBeenCalledWith('JAWA BARAT', expect.any(Object))
      expect(earthquakeService.listLatest).toHaveBeenCalledWith(expect.any(Object))
      expect(hospitalService.listHospitals).toHaveBeenCalledWith('3273', 1, expect.any(Object))
    })

    expect(screen.getByText('PASTEUR, SUKAJADI, KOTA BANDUNG, JAWA BARAT')).toBeInTheDocument()
    expect(screen.getByText('RS Hasan Sadikin')).toBeInTheDocument()
    expect(screen.getByText('Selatan Jawa Barat')).toBeInTheDocument()
  })

  it('isolates feature failure and keeps other panels functional', async () => {
    warningService.listWarnings.mockRejectedValueOnce(new Error('Peringatan gagal dimuat'))

    const user = userEvent.setup()
    renderDashboard()

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Provinsi|Province/i), '32')
    await waitFor(() => expect(screen.getByRole('option', { name: 'KOTA BANDUNG' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kabupaten\/Kota|City or regency/i), '3273')
    await waitFor(() => expect(screen.getByRole('option', { name: 'SUKAJADI' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kecamatan|District/i), '3273010')
    await waitFor(() => expect(screen.getByRole('option', { name: 'PASTEUR' })).toBeInTheDocument())
    await user.selectOptions(screen.getByLabelText(/Kelurahan\/Desa|Village/i), '3273010100')

    await waitFor(() => {
      expect(screen.getByText('Peringatan gagal dimuat')).toBeInTheDocument()
    })

    // Other panels still rendered
    expect(screen.getByText('RS Hasan Sadikin')).toBeInTheDocument()
    expect(screen.getByText('Cerah')).toBeInTheDocument()
    expect(screen.getByText('Selatan Jawa Barat')).toBeInTheDocument()
  })
})
