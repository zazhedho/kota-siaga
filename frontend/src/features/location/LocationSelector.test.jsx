import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LocationSelector } from './LocationSelector'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'
import * as locationService from './locationService'

vi.mock('./locationService', () => ({
  listProvinces: vi.fn(),
  listCities: vi.fn(),
  listDistricts: vi.fn(),
  listVillages: vi.fn(),
  searchLocations: vi.fn(),
  resolveLocation: vi.fn(),
}))

function renderComponent(props = {}) {
  return render(
    <LocaleProvider>
      <LocationSelector {...props} />
    </LocaleProvider>,
  )
}

async function chooseOption(user, label, optionName) {
  const input = screen.getByRole('combobox', { name: label })
  await waitFor(() => expect(input).toBeEnabled())
  await user.click(input)
  await user.click(await screen.findByRole('option', { name: optionName }))
}

describe('LocationSelector', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    locationService.listProvinces.mockResolvedValue({
      rows: [
        { id: '32', name: 'JAWA BARAT' },
        { id: '31', name: 'DKI JAKARTA' },
      ],
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
    locationService.searchLocations.mockResolvedValue([
      {
        id: '3273011001',
        code: '32.73.01.1001',
        name: 'PASTEUR',
        level: 'village',
        hierarchy: 'PASTEUR — SUKAJADI, KOTA BANDUNG, JAWA BARAT',
      },
    ])
    locationService.resolveLocation.mockResolvedValue({
      province: { id: '32', name: 'JAWA BARAT' },
      city: { id: '3273', name: 'KOTA BANDUNG' },
      district: { id: '3273010', name: 'SUKAJADI' },
      village: { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' },
    })
  })

  it('loads provinces on mount and disables dependent combo boxes', async () => {
    renderComponent()

    await waitFor(() => {
      expect(locationService.listProvinces).toHaveBeenCalled()
    })

    expect(screen.getByRole('combobox', { name: /City or regency|Kabupaten\/Kota/i })).toBeDisabled()
    expect(screen.getByRole('combobox', { name: /District|Kecamatan/i })).toBeDisabled()
    expect(screen.getByRole('combobox', { name: /Village|Kelurahan\/Desa/i })).toBeDisabled()
  })

  it('cascades selection through all levels and calls onComplete', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()

    renderComponent({ onComplete })

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')

    expect(onComplete).toHaveBeenCalledWith({
      province: { id: '32', name: 'JAWA BARAT' },
      city: { id: '3273', name: 'KOTA BANDUNG' },
      district: { id: '3273010', name: 'SUKAJADI' },
      village: { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' },
      adm4: '32.73.01.1001',
    })
  })

  it('searches villages directly and fills the complete location path', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()

    renderComponent({ onComplete })

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await user.click(screen.getByRole('button', { name: /Cari wilayah langsung|Search location directly/i }))

    const searchInput = screen.getByRole('combobox', { name: /Cari Kecamatan atau Kelurahan\/Desa|Search District or Village/i })
    await user.type(searchInput, 'pas')

    await waitFor(() => {
      expect(locationService.searchLocations).toHaveBeenCalledWith('pas', 10, expect.any(AbortSignal))
    })
    await user.click(await screen.findByRole('option', { name: /PASTEUR/ }))

    await waitFor(() => {
      expect(locationService.resolveLocation).toHaveBeenCalledWith('32.73.01.1001', expect.any(AbortSignal))
      expect(screen.getByRole('combobox', { name: /Provinsi|Province/i })).toHaveValue('JAWA BARAT')
      expect(screen.getByRole('combobox', { name: /Kabupaten\/Kota|City or regency/i })).toHaveValue('KOTA BANDUNG')
      expect(screen.getByRole('combobox', { name: /Kecamatan|District/i })).toHaveValue('SUKAJADI')
      expect(screen.getByRole('combobox', { name: /Kelurahan\/Desa|Village/i })).toHaveValue('PASTEUR')
    })
    expect(screen.queryByText('32.73.01.1001')).not.toBeInTheDocument()

    expect(onComplete).toHaveBeenCalledTimes(1)
    expect(onComplete).toHaveBeenCalledWith(expect.objectContaining({ village: { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' } }))
  })

  it('shows district hierarchy and loads villages before completing the location', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()
    const districtPath = {
      level: 'district',
      province: { id: '32', name: 'JAWA BARAT' },
      city: { id: '3273', name: 'KOTA BANDUNG' },
      district: { id: '3273010', name: 'SUKAJADI' },
      village: null,
    }
    locationService.searchLocations.mockResolvedValueOnce([
      {
        id: '327301',
        code: '32.73.01',
        name: 'SUKAJADI',
        level: 'district',
        hierarchy: 'SUKAJADI — KOTA BANDUNG, JAWA BARAT',
      },
    ])
    locationService.resolveLocation.mockResolvedValueOnce(districtPath)

    renderComponent({ onComplete })

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await user.click(screen.getByRole('button', { name: /Cari wilayah langsung|Search location directly/i }))
    const searchInput = screen.getByRole('combobox', { name: /Cari Kecamatan atau Kelurahan\/Desa|Search District or Village/i })
    await user.type(searchInput, 'suk')

    await waitFor(() => expect(locationService.searchLocations).toHaveBeenCalledWith('suk', 10, expect.any(AbortSignal)))
    expect(screen.getByRole('option', { name: 'SUKAJADI — KOTA BANDUNG, JAWA BARAT' })).toBeInTheDocument()
    expect(screen.queryByText('32.73.01')).not.toBeInTheDocument()

    await user.click(screen.getByRole('option', { name: 'SUKAJADI — KOTA BANDUNG, JAWA BARAT' }))

    await waitFor(() => {
      expect(locationService.resolveLocation).toHaveBeenCalledWith('32.73.01', expect.any(AbortSignal))
      expect(screen.getByRole('combobox', { name: /Provinsi|Province/i })).toHaveValue('JAWA BARAT')
      expect(screen.getByRole('combobox', { name: /Kabupaten\/Kota|City or regency/i })).toHaveValue('KOTA BANDUNG')
      expect(screen.getByRole('combobox', { name: /Kecamatan|District/i })).toHaveValue('SUKAJADI')
      expect(screen.getByRole('combobox', { name: /Kelurahan\/Desa|Village/i })).toHaveValue('')
      expect(locationService.listVillages).toHaveBeenCalledWith('3273010', expect.any(AbortSignal))
    })
    expect(onComplete).toHaveBeenLastCalledWith(null)

    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')
    expect(onComplete).toHaveBeenLastCalledWith(expect.objectContaining({ adm4: '32.73.01.1001' }))
  })

  it('cancels a stale direct-search request', async () => {
    const user = userEvent.setup()
    let firstSignal
    locationService.searchLocations.mockImplementationOnce((_query, _limit, signal) => {
      firstSignal = signal
      return new Promise(() => {})
    })

    renderComponent()
    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await user.click(screen.getByRole('button', { name: /Cari wilayah langsung|Search location directly/i }))
    const searchInput = screen.getByRole('combobox', { name: /Cari Kecamatan atau Kelurahan\/Desa|Search District or Village/i })
    await user.type(searchInput, 'pas')
    await waitFor(() => expect(locationService.searchLocations).toHaveBeenCalled())

    await user.clear(searchInput)
    await user.type(searchInput, 'suk')

    await waitFor(() => expect(firstSignal).toHaveProperty('aborted', true))
  })

  it('keeps cascading selection available after direct search is closed', async () => {
    const user = userEvent.setup()
    renderComponent()

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    const toggle = screen.getByRole('button', { name: /Cari wilayah langsung|Search location directly/i })
    await user.click(toggle)
    expect(screen.getByRole('combobox', { name: /Cari Kecamatan atau Kelurahan\/Desa|Search District or Village/i })).toBeInTheDocument()
    await user.click(toggle)
    expect(screen.queryByRole('combobox', { name: /Cari Kecamatan atau Kelurahan\/Desa|Search District or Village/i })).not.toBeInTheDocument()

    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    expect(locationService.listCities).toHaveBeenCalledWith('32', expect.any(AbortSignal))
  })

  it('clears descendants when parent selection changes', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()

    renderComponent({ onComplete })

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalled())
    await chooseOption(user, /Provinsi|Province/i, 'JAWA BARAT')
    await chooseOption(user, /Kabupaten\/Kota|City or regency/i, 'KOTA BANDUNG')
    await chooseOption(user, /Kecamatan|District/i, 'SUKAJADI')
    await chooseOption(user, /Kelurahan\/Desa|Village/i, 'PASTEUR')

    // Change city
    const city = screen.getByRole('combobox', { name: /Kabupaten\/Kota|City or regency/i })
    await user.click(city)
    await user.clear(city)

    expect(screen.getByRole('combobox', { name: /Kecamatan|District/i })).toHaveValue('')
    expect(screen.getByRole('combobox', { name: /Kelurahan\/Desa|Village/i })).toHaveValue('')
    expect(screen.getByRole('combobox', { name: /Kecamatan|District/i })).toBeDisabled()
    expect(onComplete).toHaveBeenLastCalledWith(null)
  })

  it('renders retry button on failure', async () => {
    locationService.listProvinces.mockRejectedValueOnce(new Error('Network error'))

    renderComponent()

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument()
    })

    const retryBtn = screen.getByRole('button', { name: /Coba Lagi|Retry/i })
    expect(retryBtn).toBeInTheDocument()

    locationService.listProvinces.mockResolvedValueOnce({
      rows: [{ id: '32', name: 'JAWA BARAT' }],
    })

    const user = userEvent.setup()
    await user.click(retryBtn)

    await waitFor(() => expect(locationService.listProvinces).toHaveBeenCalledTimes(2))
    const province = screen.getByRole('combobox', { name: /Provinsi|Province/i })
    await user.click(province)
    expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
  })

  it('restores a stored location and notifies the dashboard after all levels load', async () => {
    const onComplete = vi.fn()

    renderComponent({
      initialLocation: {
        provinceId: '32',
        cityId: '3273',
        districtId: '3273010',
        villageId: '3273010100',
        adm4: '32.73.01.1001',
      },
      onComplete,
    })

    await waitFor(() => {
      expect(onComplete).toHaveBeenCalledWith({
        province: { id: '32', name: 'JAWA BARAT' },
        city: { id: '3273', name: 'KOTA BANDUNG' },
        district: { id: '3273010', name: 'SUKAJADI' },
        village: { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' },
        adm4: '32.73.01.1001',
      })
    })
  })
})
