const API_BASE = import.meta.env.VITE_API_URL || '';

export interface ApiError {
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

async function apiUpload(path: string, formData: FormData): Promise<unknown> {
	const res = await fetch(`${API_BASE}${path}`, {
		method: 'POST',
		credentials: 'include',
		body: formData,
		// Do NOT set Content-Type — browser sets multipart boundary automatically
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
	store: (key: string, familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'POST', body: JSON.stringify({ apiKey: key, familyCode }) }),
	update: (key: string, familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'PUT', body: JSON.stringify({ apiKey: key, familyCode }) }),
	delete: (familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'DELETE', body: JSON.stringify({ familyCode }) }),
};

// Children
export const children = {
	list: () => apiFetch('/api/children'),
	create: (name: string) =>
		apiFetch('/api/children', { method: 'POST', body: JSON.stringify({ name }) }),
	generateInvite: (id: string) =>
		apiFetch(`/api/children/${id}/invite`, { method: 'POST' }),
	loginAs: (id: string) =>
		apiFetch(`/api/children/${id}/login-as`, { method: 'POST' }),
};

// Invite
export const invite = {
	activate: (token: string, name: string, pin: string) =>
		apiFetch(`/api/invite/${token}/activate`, {
			method: 'POST',
			body: JSON.stringify({ name, pin }),
		}),
};

// Content
export const content = {
	subjects: () => apiFetch('/api/subjects'),
	topics: (subjectSlug: string) => apiFetch(`/api/subjects/${subjectSlug}/topics`),
	exercises: (subjectSlug: string, topicSlug: string) =>
		apiFetch(`/api/topics/${subjectSlug}/${topicSlug}/exercises`),
};

// Sessions
export const sessions = {
	create: (exerciseId: string) =>
		apiFetch('/api/sessions', {
			method: 'POST',
			body: JSON.stringify({ exerciseId }),
		}),
	createChallenge: (challengeExerciseId: string) =>
		apiFetch('/api/sessions', {
			method: 'POST',
			body: JSON.stringify({ challengeExerciseId }),
		}),
	get: (sessionId: string) => apiFetch(`/api/sessions/${sessionId}`),
	end: (sessionId: string) =>
		apiFetch(`/api/sessions/${sessionId}/end`, { method: 'POST' }),
	messages: (sessionId: string) => apiFetch(`/api/sessions/${sessionId}/messages`),
};

// Reviews (spaced repetition)
export const reviews = {
	due: () => apiFetch('/api/reviews/due'),
};

// Progress
export const progress = {
	mine: () => apiFetch('/api/progress'),
	child: (studentId: string) => apiFetch(`/api/children/${studentId}/progress`),
	childSessions: (studentId: string) =>
		apiFetch(`/api/children/${studentId}/progress/sessions`),
};

// Challenges
export const challenges = {
	list: () => apiFetch('/api/challenges'),
	get: (id: string) => apiFetch(`/api/challenges/${id}`),
	create: (files: File[], text = '') => {
		const form = new FormData();
		files.forEach((f) => form.append('files', f));
		if (text.trim()) form.append('text', text.trim());
		return apiUpload('/api/challenges', form);
	},
	publish: (id: string, title: string) =>
		apiFetch(`/api/challenges/${id}/publish`, {
			method: 'PATCH',
			body: JSON.stringify({ title }),
		}),
	delete: (id: string) => apiFetch(`/api/challenges/${id}`, { method: 'DELETE' }),
};

// Telegram
export const telegram = {
	createLinkCode: () =>
		apiFetch('/api/telegram/link-code', { method: 'POST' }) as Promise<{
			code: string;
			link: string;
		}>,
};

// Device / family profiles
export const device = {
	unlock: (code: string) =>
		apiFetch('/api/device/unlock', { method: 'POST', body: JSON.stringify({ code }) }),
	profiles: () =>
		apiFetch('/api/profiles') as Promise<{ id: string; name: string; role: 'parent' | 'child' }[]>,
	profileLogin: (id: string, role: 'parent' | 'child') =>
		apiFetch('/api/profile/login', { method: 'POST', body: JSON.stringify({ id, role }) }) as Promise<{
			role: string;
			name: string;
		}>,
	setCode: (newCode: string, currentCode?: string) =>
		apiFetch('/api/device/set-code', {
			method: 'POST',
			body: JSON.stringify({ newCode, currentCode }),
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
