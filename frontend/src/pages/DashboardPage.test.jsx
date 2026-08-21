import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
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

async function chooseOption(user, label, optionName) {
  const input = screen.getByRole('combobox', { name: label })
  await waitFor(() => expect(input).toBeEnabled())
  await user.click(input)
  await user.click(await screen.findByRole('option', { name: optionName }))
}

function deferred() {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

describe('DashboardPage', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

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

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')

    await waitFor(() => {
      expect(weatherService.getForecast).toHaveBeenCalledWith('32.73.01.1001', expect.any(Object))
      expect(warningService.listWarnings).toHaveBeenCalledWith('JAWA BARAT', expect.any(Object))
      expect(earthquakeService.listLatest).toHaveBeenCalledWith(expect.any(Object))
      expect(hospitalService.listHospitals).toHaveBeenCalledWith('3273', 1, '', expect.any(Object))
    })

    expect(screen.getByText('PASTEUR, SUKAJADI, KOTA BANDUNG, JAWA BARAT')).toBeInTheDocument()
    expect(screen.getByText('RS Hasan Sadikin')).toBeInTheDocument()
    expect(screen.getByText('Selatan Jawa Barat')).toBeInTheDocument()
  })

  it('isolates feature failure and keeps other panels functional', async () => {
    warningService.listWarnings.mockRejectedValueOnce(new Error('Peringatan gagal dimuat'))

    const user = userEvent.setup()
    renderDashboard()

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')

    await waitFor(() => {
      expect(screen.getByText('Peringatan gagal dimuat')).toBeInTheDocument()
    })

    // Other panels still rendered
    expect(screen.getByText('RS Hasan Sadikin')).toBeInTheDocument()
    expect(screen.getByText('Cerah')).toBeInTheDocument()
    expect(screen.getByText('Selatan Jawa Barat')).toBeInTheDocument()
  })

  it('debounces hospital search and preserves committed search for load more', async () => {
    const user = userEvent.setup()
    hospitalService.listHospitals
      .mockResolvedValueOnce({
        rows: [{ id: 'h1', name: 'RS Hasan Sadikin' }],
        total: 1,
        page: 1,
        nextPage: false,
      })
      .mockResolvedValueOnce({
        rows: [{ id: 'h2', name: 'RS Hasan Search Result' }],
        total: 2,
        page: 1,
        nextPage: true,
      })
      .mockResolvedValueOnce({
        rows: [{ id: 'h3', name: 'RS Hasan Search Result 2' }],
        total: 2,
        page: 2,
        nextPage: false,
      })

    renderDashboard()
    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenCalledWith('3273', 1, '', expect.any(Object))
    })

    const initialSignal = hospitalService.listHospitals.mock.calls[0][3]
    vi.useFakeTimers()
    fireEvent.change(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i }), {
      target: { value: '  Hasan Sadikin  ' },
    })
    act(() => vi.advanceTimersByTime(299))
    expect(hospitalService.listHospitals).toHaveBeenCalledTimes(1)
    act(() => vi.advanceTimersByTime(1))
    vi.useRealTimers()

    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith(
        '3273',
        1,
        'Hasan Sadikin',
        expect.any(Object),
      )
    })
    expect(hospitalService.listHospitals.mock.calls[1][3]).toBe(initialSignal)
    expect(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })).toHaveValue('Hasan Sadikin')

    fireEvent.click(screen.getByRole('button', { name: /Muat Lebih Banyak|Load More/i }))
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith(
        '3273',
        2,
        'Hasan Sadikin',
        expect.any(Object),
      )
    })
    expect(hospitalService.listHospitals.mock.calls[2][3]).toBe(initialSignal)
  })

  it('clears hospital search when the monitored location changes', async () => {
    const user = userEvent.setup()
    locationService.listProvinces.mockResolvedValue({
      rows: [
        { id: '32', name: 'JAWA BARAT' },
        { id: '33', name: 'JAWA TENGAH' },
      ],
    })
    locationService.listCities.mockImplementation((provinceId) => Promise.resolve({
      rows: provinceId === '33'
        ? [{ id: '3374', name: 'KOTA SEMARANG' }]
        : [{ id: '3273', name: 'KOTA BANDUNG' }],
    }))
    locationService.listDistricts.mockImplementation((cityId) => Promise.resolve({
      rows: cityId === '3374'
        ? [{ id: '3374010', name: 'SEMARANG TENGAH' }]
        : [{ id: '3273010', name: 'SUKAJADI' }],
    }))
    locationService.listVillages.mockImplementation((districtId) => Promise.resolve({
      rows: districtId === '3374010'
        ? [{ id: '3374010100', name: 'BRUMBUNG', code: '33.74.01.1001' }]
        : [{ id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' }],
    }))

    renderDashboard()
    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3273', 1, '', expect.any(Object))
    })

    vi.useFakeTimers()
    fireEvent.change(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i }), {
      target: { value: 'Hasan' },
    })
    act(() => vi.advanceTimersByTime(300))
    vi.useRealTimers()
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3273', 1, 'Hasan', expect.any(Object))
    })

    const provinceInput = screen.getByRole('combobox', { name: /Provinsi|Province/i })
    await user.click(provinceInput)
    await user.clear(provinceInput)
    await user.click(await screen.findByRole('option', { name: 'JAWA TENGAH' }))
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA SEMARANG')
    await chooseOption(user, /Kecamatan|District/i, 'SEMARANG TENGAH')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'BRUMBUNG')

    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3374', 1, '', expect.any(Object))
    })
    expect(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i })).toHaveValue('')
  })

  it('clears stale hospital state before a pending request for a new monitored village', async () => {
    const user = userEvent.setup()
    const newHospitalRequest = deferred()
    locationService.listVillages.mockResolvedValue({
      rows: [
        { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' },
        { id: '3273010101', name: 'CIPAGANTI', code: '32.73.01.1002' },
      ],
    })
    hospitalService.listHospitals
      .mockResolvedValueOnce({
        rows: [{ id: 'old', name: 'Old Hospital' }],
        total: 25,
        page: 1,
        nextPage: true,
      })
      .mockReturnValueOnce(newHospitalRequest.promise)

    renderDashboard()
    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Village|Desa/i, 'PASTEUR')
    await waitFor(() => {
      expect(screen.getByText('Old Hospital')).toBeInTheDocument()
      expect(screen.getByText('Menampilkan 1 dari 25 fasilitas kesehatan')).toBeInTheDocument()
    })

    const villageInput = screen.getByRole('combobox', { name: /Village|Desa/i })
    await user.click(villageInput)
    fireEvent.change(villageInput, { target: { value: 'CIPAGANTI' } })
    await user.click(await screen.findByRole('option', { name: 'CIPAGANTI' }))
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3273', 1, '', expect.any(Object))
    })

    expect(screen.queryByText('Old Hospital')).not.toBeInTheDocument()
    expect(screen.queryByText('Menampilkan 1 dari 25 fasilitas kesehatan')).not.toBeInTheDocument()

    newHospitalRequest.resolve({
      rows: [{ id: 'new', name: 'New Hospital' }],
      total: 1,
      page: 1,
      nextPage: false,
    })
    await waitFor(() => expect(screen.getByText('New Hospital')).toBeInTheDocument())
  })

  it('ignores a stale hospital response from an older search', async () => {
    const user = userEvent.setup()
    const oldRequest = deferred()
    const newRequest = deferred()
    const initialResult = {
      rows: [{ id: 'h1', name: 'RS Hasan Sadikin' }],
      total: 1,
      page: 1,
      nextPage: false,
    }
    hospitalService.listHospitals.mockImplementation((_cityId, _page, search) => {
      if (search === 'Old') return oldRequest.promise
      if (search === 'New') return newRequest.promise
      return Promise.resolve(initialResult)
    })

    renderDashboard()
    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')
    await waitFor(() => {
      expect(hospitalService.listHospitals).toHaveBeenCalledWith('3273', 1, '', expect.any(Object))
    })

    vi.useFakeTimers()
    fireEvent.change(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i }), {
      target: { value: 'Old' },
    })
    act(() => vi.advanceTimersByTime(300))
    expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3273', 1, 'Old', expect.any(Object))

    fireEvent.change(screen.getByRole('textbox', { name: /Cari rumah sakit|Search hospitals/i }), {
      target: { value: 'New' },
    })
    act(() => vi.advanceTimersByTime(300))
    expect(hospitalService.listHospitals).toHaveBeenLastCalledWith('3273', 1, 'New', expect.any(Object))
    vi.useRealTimers()

    newRequest.resolve({
      rows: [{ id: 'new', name: 'New Hospital' }],
      total: 1,
      page: 1,
      nextPage: false,
    })
    await waitFor(() => expect(screen.getByText('New Hospital')).toBeInTheDocument())

    oldRequest.resolve({
      rows: [{ id: 'old', name: 'Old Hospital' }],
      total: 1,
      page: 1,
      nextPage: false,
    })
    await act(async () => {
      await oldRequest.promise
    })
    expect(screen.queryByText('Old Hospital')).not.toBeInTheDocument()
  })
})
