import { LocaleProvider } from './shared/i18n/LocaleProvider'
import { AppShell } from './components/layout/AppShell'
import { DashboardPage } from './pages/DashboardPage'

export default function App() {
  return (
    <LocaleProvider>
      <AppShell>
        <DashboardPage />
      </AppShell>
    </LocaleProvider>
  )
}
