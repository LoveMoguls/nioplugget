const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface ApiError {
	status: number;
	data: { error?: string } | null;
}

async function apiFetch(path: string, options: RequestInit = {}): Promise<unknown> {
	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		credentials: 'include', // Send httpOnly cookies
		headers: {
			'Content-Type': 'application/json',
			...options.headers,
		},
	});
	const data = await res.json().catch(() => null);
	if (!res.ok) throw { status: res.status, data } as ApiError;
	return data;
}

export function getErrorMessage(err: unknown, fallback = 'Ett fel inträffade'): string {
	if (err && typeof err === 'object' && 'data' in err) {
		const apiErr = err as ApiError;
		if (apiErr.data && typeof apiErr.data === 'object' && 'error' in apiErr.data) {
			return apiErr.data.error || fallback;
		}
	}
	return fallback;
}

// Auth
export const auth = {
	register: (email: string, password: string, gdprConsent: boolean) =>
		apiFetch('/api/auth/register', {
			method: 'POST',
			body: JSON.stringify({ email, password, gdprConsent }),
		}),
	login: (email: string, password: string) =>
		apiFetch('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password }),
		}),
	logout: () => apiFetch('/api/auth/logout', { method: 'POST' }),
	me: () => apiFetch('/api/auth/me'),
};

// API Keys
export const apiKey = {
	get: () => apiFetch('/api/apikey'),
	store: (key: string) =>
		apiFetch('/api/apikey', { method: 'POST', body: JSON.stringify({ apiKey: key }) }),
	update: (key: string) =>
		apiFetch('/api/apikey', { method: 'PUT', body: JSON.stringify({ apiKey: key }) }),
	delete: () => apiFetch('/api/apikey', { method: 'DELETE' }),
};

// Children
export const children = {
	list: () => apiFetch('/api/children'),
	create: (name: string) =>
		apiFetch('/api/children', { method: 'POST', body: JSON.stringify({ name }) }),
	generateInvite: (id: string) =>
		apiFetch(`/api/children/${id}/invite`, { method: 'POST' }),
};

// Invite
export const invite = {
	activate: (token: string, name: string, pin: string) =>
		apiFetch(`/api/invite/${token}/activate`, {
			method: 'POST',
			body: JSON.stringify({ name, pin }),
		}),
};

// Child auth
export const childAuth = {
	names: (parentEmail: string) =>
		apiFetch(`/api/child/names?parent_email=${encodeURIComponent(parentEmail)}`),
	login: (parentEmail: string, studentName: string, pin: string) =>
		apiFetch('/api/child/login', {
			method: 'POST',
			body: JSON.stringify({ parentEmail, studentName, pin }),
		}),
};
