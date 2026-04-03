import { writable, derived } from 'svelte/store';
import { auth as authApi, getErrorMessage } from '$lib/api';

interface User {
	id: string;
	email: string;
	role: 'parent' | 'child';
}

function createAuthStore() {
	const { subscribe, set, update } = writable<User | null>(null);

	return {
		subscribe,

		// Set the user directly (e.g., after login/register)
		setUser: (user: User | null) => set(user),

		// Check auth status by calling /api/auth/me (restores session from cookie)
		checkAuth: async () => {
			try {
				const data = (await authApi.me()) as User;
				if (data && data.id) {
					set({ id: data.id, email: data.email || '', role: data.role });
				} else {
					set(null);
				}
			} catch {
				set(null);
			}
		},

		// Logout: call API, clear store
		logout: async () => {
			try {
				await authApi.logout();
			} catch (err) {
				// Ignore errors — clear local state regardless
				void getErrorMessage(err);
			}
			set(null);
		},

		// Clear store without API call
		clear: () => set(null),
	};
}

export const user = createAuthStore();
export const isLoggedIn = derived(user, ($user) => $user !== null);
export const isParent = derived(user, ($user) => $user?.role === 'parent');
export const isChild = derived(user, ($user) => $user?.role === 'child');
