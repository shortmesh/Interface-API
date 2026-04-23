import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { ConfigProvider, theme as antdTheme } from 'antd'
import App from './App'
import theme from './theme'
import { antdTheme as customAntdTheme } from './antdTheme'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter basename="/admin">
      <ThemeProvider theme={theme}>
        <ConfigProvider
          theme={{
            ...customAntdTheme,
            algorithm: antdTheme.darkAlgorithm,
          }}
        >
          <CssBaseline />
          <App />
        </ConfigProvider>
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
