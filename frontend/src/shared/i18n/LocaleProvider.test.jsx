import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LocaleProvider, useLocale } from './index'

function TestComponent() {
  const { locale, setLocale, t } = useLocale()
  return (
    <div>
      <span data-testid="current-locale">{locale}</span>
      <span data-testid="translated-title">{t('brandTitle')}</span>
      <span data-testid="interpolated">{t('showingHospitalsCount', { count: 5, total: 20 })}</span>
      <button onClick={() => setLocale('en')}>Switch to English</button>
      <button onClick={() => setLocale('id')}>Switch to Indonesian</button>
    </div>
  )
}

describe('LocaleProvider', () => {
  it('defaults to Indonesian locale and persists switch', async () => {
    const user = userEvent.setup()
    render(
      <LocaleProvider>
        <TestComponent />
      </LocaleProvider>,
    )

    expect(screen.getByTestId('current-locale')).toHaveTextContent('id')
    expect(screen.getByTestId('interpolated')).toHaveTextContent('Menampilkan 5 dari 20 fasilitas kesehatan')

    await user.click(screen.getByText('Switch to English'))

    expect(screen.getByTestId('current-locale')).toHaveTextContent('en')
    expect(localStorage.getItem('kota-siaga.locale')).toBe('en')
    expect(document.documentElement.lang).toBe('en')
    expect(screen.getByTestId('interpolated')).toHaveTextContent('Showing 5 of 20 healthcare facilities')
  })

  it('restores locale from localStorage', () => {
    localStorage.setItem('kota-siaga.locale', 'en')

    render(
      <LocaleProvider>
        <TestComponent />
      </LocaleProvider>,
    )

    expect(screen.getByTestId('current-locale')).toHaveTextContent('en')
  })
})
