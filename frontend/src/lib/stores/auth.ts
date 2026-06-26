import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import * as authApi from '$lib/api/auth';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';

function createAuthStore() {
	const stored = browser ? localStorage.getItem('token') : null;
	const { subscribe, set } = writable<string | null>(stored);

	return {
		subscribe,

		async login(username: string, password: string): Promise<void> {
			const res = await authApi.login(username, password);
			localStorage.setItem('token', res.token);
			set(res.token);
		},

		async register(username: string, password: string): Promise<void> {
			await authApi.register(username, password);
		},

		logout() {
			localStorage.removeItem('token');
			set(null);
			goto(resolve('/login'));
		},

		isAuthenticated(): boolean {
			let token: string | null = null;
			subscribe((v) => (token = v))();
			return token !== null;
		}
	};
}

export const auth = createAuthStore();
