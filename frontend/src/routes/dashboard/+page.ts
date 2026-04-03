import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { get } from 'svelte/store';
import { user } from '$lib/stores/auth';

export const load: PageLoad = async () => {
	// Check if user is logged in as parent — if not, redirect to login
	// Note: user store is populated by checkAuth() in +layout.svelte on mount
	// This load function runs on navigation, but store may not be populated yet for SSR
	// We rely on the page component to handle the redirect after mount
	return {};
};
