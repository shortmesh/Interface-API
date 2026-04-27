// Scope management utilities

const SCOPES_KEY = 'shortmesh_scopes'

export const saveScopes = (scopes) => {
  localStorage.setItem(SCOPES_KEY, JSON.stringify(scopes))
}

export const getScopes = () => {
  const scopes = localStorage.getItem(SCOPES_KEY)
  const parsed = scopes ? JSON.parse(scopes) : []
  return parsed
}

export const clearScopes = () => {
  localStorage.removeItem(SCOPES_KEY)
}

/**
 * Check if user has a specific scope
 * Supports wildcard matching:
 * - "*" matches everything (full access)
 * - "tokens:*" matches all token operations
 * - "tokens:write:*" matches all token write operations
 */
export const hasScope = (requiredScope) => {
  const userScopes = getScopes()
  const hasPermission = userScopes.some((scope) => {
    // Global wildcard - full access to everything
    if (scope === '*') {
      return true
    }
    
    // Exact match
    if (scope === requiredScope) {
      return true
    }
    
    // Wildcard match - e.g., "tokens:*" matches "tokens:read:*" or "tokens:write:create"
    if (scope.endsWith(':*')) {
      const prefix = scope.slice(0, -1) // Remove the *
      return requiredScope.startsWith(prefix)
    }
    
    return false
  })
  
  return hasPermission
}

/**
 * Check if user has any of the required scopes
 */
export const hasAnyScope = (...requiredScopes) => {
  return requiredScopes.some((scope) => hasScope(scope))
}

/**
 * Check if user has all of the required scopes
 */
export const hasAllScopes = (...requiredScopes) => {
  return requiredScopes.every((scope) => hasScope(scope))
}
