import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from '@/components/AppLayout'
import { BacktestPage } from '@/pages/BacktestPage'
import { IngestPage } from '@/pages/IngestPage'
import { InspectPage } from '@/pages/InspectPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<InspectPage />} />
          <Route path="/ingest" element={<IngestPage />} />
          <Route path="/backtest" element={<BacktestPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
