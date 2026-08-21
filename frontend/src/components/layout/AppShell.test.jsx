import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AppShell } from './AppShell'
import { LocaleProvider } from '../../shared/i18n/LocaleProvider'

describe('AppShell', () => {
  it('keeps the live label separate from the truncated tagline', () => {
    render(
      <LocaleProvider>
        <AppShell>Content</AppShell>
      </LocaleProvider>,
    )

    expect(screen.getByText('Portal Langsung')).toHaveClass('ks-live-label')
    expect(document.querySelector('.ks-live-dot')).toHaveClass('ks-live-dot')
    expect(document.querySelector('.ks-brand-tagline')).toBeInTheDocument()
  })

  it('translates the footer when the locale changes', async () => {
    const user = userEvent.setup()

    render(
      <LocaleProvider>
        <AppShell>Content</AppShell>
      </LocaleProvider>,
    )

    await user.click(screen.getByRole('button', { name: 'EN' }))

    expect(screen.getByText(/Weather and earthquake data from BMKG/)).toBeInTheDocument()
    expect(screen.queryByText(/Data cuaca & gempa bersumber/)).not.toBeInTheDocument()
  })
})
