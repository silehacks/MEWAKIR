const API_BASE = '/api';

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || 'Request failed');
  }
  const contentType = response.headers.get('Content-Type') || '';
  if (contentType.includes('application/json')) {
    return response.json();
  }
  return response.text();
}

export const api = {
  login: (name) => request('/login', { method: 'POST', body: JSON.stringify({ name }) }),
  createMeeting: () => request('/meetings', { method: 'POST' }),
  joinMeeting: (meetingId, password) =>
    request('/meetings/join', {
      method: 'POST',
      body: JSON.stringify({ meetingId, password }),
    }),
};
