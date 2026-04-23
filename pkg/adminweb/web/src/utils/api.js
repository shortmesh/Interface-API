export const apiCall = async (url, options = {}) => {
  try {
    const response = await fetch(url, options)

    if (response.status === 401) {
      window.location.href = '/admin/login'
      return null
    }

    return response
  } catch (error) {
    console.error(`Network error for ${url}:`, error)
    throw new Error('Unable to connect to the server. Please check your connection.')
  }
}

export const safeJsonParse = async (response) => {
  try {
    return await response.json()
  } catch (error) {
    console.error('Failed to parse JSON response:', error)
    return { error: 'Invalid server response' }
  }
}

export const maskString = (str) => {
  if (!str) return ''
  if (str.length <= 4) return '****'
  const visible = str.slice(-4)
  return '*'.repeat(str.length - 4) + visible
}

export const formatDate = (dateString) => {
  if (!dateString) return '-'
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return dateString
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
  } catch {
    return dateString
  }
}

export const copyToClipboard = (text) => {
  return navigator.clipboard.writeText(text)
}
