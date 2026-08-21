import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ComboBox } from './ComboBox'

const options = [
  { id: '32', name: 'JAWA BARAT' },
  { id: '31', name: 'DKI JAKARTA' },
]

function renderCombo(props = {}) {
  return render(
    <>
      <label htmlFor="province-select">Province</label>
      <ComboBox
        id="province-select"
        options={options}
        placeholder="Select Province..."
        noResultsLabel="No locations found"
        {...props}
      />
    </>,
  )
}

describe('ComboBox', () => {
  it('filters options and selects with the keyboard', async () => {
    const onChange = vi.fn()
    const user = userEvent.setup()

    renderCombo({ onChange })
    const input = screen.getByRole('combobox', { name: 'Province' })

    await user.click(input)
    await user.type(input, 'JAWA')

    expect(screen.getByRole('option', { name: 'JAWA BARAT' })).toBeInTheDocument()
    expect(screen.queryByRole('option', { name: 'DKI JAKARTA' })).not.toBeInTheDocument()

    await user.keyboard('{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenCalledWith('32')
    expect(input).toHaveValue('JAWA BARAT')
  })

  it('restores the selected label when search is cancelled', async () => {
    const user = userEvent.setup()

    renderCombo({ value: '32' })
    const input = screen.getByRole('combobox', { name: 'Province' })

    await user.click(input)
    await user.clear(input)
    await user.type(input, 'DKI')
    await user.keyboard('{Escape}')

    expect(input).toHaveValue('JAWA BARAT')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('shows no-results feedback and respects disabled state', async () => {
    const user = userEvent.setup()

    const { unmount } = renderCombo({ disabled: true })
    const input = screen.getByRole('combobox', { name: 'Province' })
    expect(input).toBeDisabled()

    unmount()
    renderCombo({ noResultsLabel: 'Wilayah tidak ditemukan' })
    const secondInput = screen.getByRole('combobox', { name: 'Province' })
    await user.click(secondInput)
    await user.type(secondInput, 'XYZ')

    expect(screen.getByText('Wilayah tidak ditemukan')).toBeInTheDocument()
  })

  it('reports remote queries and shows loading feedback', async () => {
    const onQueryChange = vi.fn()
    const user = userEvent.setup()

    renderCombo({
      ariaBusy: true,
      loadingLabel: 'Searching locations...',
      onQueryChange,
    })
    const input = screen.getByRole('combobox', { name: 'Province' })

    await user.click(input)
    await user.type(input, 'pas')

    expect(onQueryChange).toHaveBeenLastCalledWith('pas')
    expect(screen.getByText('Searching locations...')).toBeInTheDocument()
  })
})
