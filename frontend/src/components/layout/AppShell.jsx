import { AppHeader } from './AppHeader'
import { useLocale } from '../../shared/i18n'

export function AppShell({ children }) {
  const { t } = useLocale()

  return (
    <div className="d-flex flex-column min-vh-100">
      <AppHeader />
      <main className="container-fluid container-xl py-4 flex-grow-1">
        {children}
      </main>
      <footer className="py-3 bg-white border-top text-center text-secondary small">
        <div className="container-fluid container-xl">
          <span>{t('footerAttribution')}</span>
        </div>
      </footer>
    </div>
  )
}
