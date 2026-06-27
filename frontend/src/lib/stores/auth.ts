import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import * as authApi from '$lib/api/auth';
import { getToken, setToken, clearToken } from '$lib/api/client';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';

function createAuthStore() {
	const stored = browser ? getToken() : null;
	const { subscribe, set } = writable<string | null>(stored);

	return {
		subscribe,

		async login(username: string, password: string): Promise<void> {
			const res = await authApi.login(username, password);
			setToken(res.token);
			set(res.token);
		},

		async register(username: string, password: string): Promise<void> {
			await authApi.register(username, password);
			await this.login(username, password);
		},

		logout() {
			clearToken();
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
