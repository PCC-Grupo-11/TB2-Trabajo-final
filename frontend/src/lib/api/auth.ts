import { apiClient } from './client';
import type {
	LoginRequest,
	LoginResponse,
	RegisterRequest,
	RegisterResponse
} from '$lib/types/auth';

export async function login(username: string, password: string): Promise<LoginResponse> {
	return apiClient<LoginResponse>('/auth/login', {
		method: 'POST',
		body: JSON.stringify({ username, password } satisfies LoginRequest)
	});
}

export async function register(username: string, password: string): Promise<RegisterResponse> {
	return apiClient<RegisterResponse>('/auth/register', {
		method: 'POST',
		body: JSON.stringify({ username, password } satisfies RegisterRequest)
	});
}
