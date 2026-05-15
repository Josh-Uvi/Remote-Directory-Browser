// Shared application utilities

/**
 * API helper for making authenticated requests
 */
async function apiCall(endpoint, options = {}) {
    const response = await fetch(endpoint, {
        headers: {
            'Content-Type': 'application/json',
            ...options.headers,
        },
        ...options,
    });
    return response;
}

/**
 * Check if user is authenticated
 */
async function isAuthenticated() {
    try {
        const response = await fetch('/api/list?path=/', {
            method: 'GET',
        });
        return response.ok;
    } catch (error) {
        return false;
    }
}

/**
 * HTML escape for XSS prevention
 */
function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;',
    };
    return text.replace(/[&<>"']/g, m => map[m]);
}
