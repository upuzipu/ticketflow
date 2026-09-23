import { api } from '@/shared/api/client';
import type { LoginRequest, Me, RegisterRequest, Registered, TokenPair } from '@/shared/api/types';

export function registerUser(body: RegisterRequest): Promise<Registered> {
  return api.post('/auth/register', body);
}

export function loginUser(body: LoginRequest): Promise<TokenPair> {
  return api.post('/auth/login', body);
}

export function logoutUser(refreshToken: string): Promise<void> {
  return api.post('/auth/logout', { refresh_token: refreshToken });
}

export function fetchMe(): Promise<Me> {
  return api.get('/users/me');
}
