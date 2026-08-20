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
}))

function renderComponent(props = {}) {
  return render(
    <LocaleProvider>
      <LocationSelector {...props} />
    </LocaleProvider>,
  )
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
  })

  it('loads provinces on mount and disables dependent selects', async () => {
    renderComponent()

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    })

    expect(screen.getByLabelText(/City or regency|Kabupaten\/Kota/i)).toBeDisabled()
    expect(screen.getByLabelText(/District|Kecamatan/i)).toBeDisabled()
    expect(screen.getByLabelText(/Village|Kelurahan\/Desa/i)).toBeDisabled()
  })

  it('cascades selection through all levels and calls onComplete', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()

    renderComponent({ onComplete })

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Provinsi|Province/i), '32')

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'KOTA BANDUNG' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Kabupaten\/Kota|City or regency/i), '3273')

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'SUKAJADI' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Kecamatan|District/i), '3273010')

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'PASTEUR' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/Kelurahan\/Desa|Village/i), '3273010100')

    expect(onComplete).toHaveBeenCalledWith({
      province: { id: '32', name: 'JAWA BARAT' },
      city: { id: '3273', name: 'KOTA BANDUNG' },
      district: { id: '3273010', name: 'SUKAJADI' },
      village: { id: '3273010100', name: 'PASTEUR', code: '32.73.01.1001' },
      adm4: '32.73.01.1001',
    })
  })

  it('clears descendants when parent selection changes', async () => {
    const onComplete = vi.fn()
    const user = userEvent.setup()

    renderComponent({ onComplete })

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

    // Change city
    await user.selectOptions(screen.getByLabelText(/Kabupaten\/Kota|City or regency/i), '')

    expect(screen.getByLabelText(/Kecamatan|District/i)).toHaveValue('')
    expect(screen.getByLabelText(/Kelurahan\/Desa|Village/i)).toHaveValue('')
    expect(screen.getByLabelText(/Kecamatan|District/i)).toBeDisabled()
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

    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    })
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
