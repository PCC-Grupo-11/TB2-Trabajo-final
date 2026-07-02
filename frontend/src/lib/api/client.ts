import type { ApiError } from '$lib/types/auth';

const BASE_URL = '/api/v1';

export class ApiRequestError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.name = 'ApiRequestError';
		this.status = status;
	}
}

export function getToken(): string | null {
	if (typeof window === 'undefined') return null;
	const token = localStorage.getItem('token');
	if (!token) return null;
	try {
		const payload = JSON.parse(atob(token.split('.')[1]));
		if (payload.exp && payload.exp * 1000 < Date.now()) {
			clearToken();
			return null;
		}
	} catch {
		clearToken();
		return null;
	}
	return token;
}

export function setToken(token: string): void {
	localStorage.setItem('token', token);
}

export function clearToken(): void {
	localStorage.removeItem('token');
}

export async function apiClient<T>(path: string, options?: RequestInit): Promise<T> {
	const token = getToken();

	const res = await fetch(`${BASE_URL}${path}`, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...options?.headers
		}
	});

	if (!res.ok) {
		let body: ApiError;
		try {
			body = await res.json();
		} catch {
			throw new ApiRequestError(res.statusText, res.status);
		}
		throw new ApiRequestError(body.error, res.status);
	}

	return res.json() as Promise<T>;
}
