import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap, catchError, of, firstValueFrom } from 'rxjs';
import { environment } from '@env/environment';

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: 'user' | 'admin';
  createdAt: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly API_URL = environment.apiUrl;

  private userSignal = signal<User | null>(null);
  private loadingSignal = signal(true);
  private initialized = false;

  readonly user = this.userSignal.asReadonly();
  readonly isAuthenticated = computed(() => this.userSignal() !== null);
  readonly isAdmin = computed(() => this.userSignal()?.role === 'admin');
  readonly loading = this.loadingSignal.asReadonly();

  constructor(
    private http: HttpClient,
    private router: Router
  ) {}

  async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }
    this.initialized = true;

    const token = this.getAccessToken();
    console.log('[AuthService] Initialize - token exists:', !!token);

    if (!token) {
      this.loadingSignal.set(false);
      return;
    }

    try {
      const user = await firstValueFrom(
        this.http.get<User>(`${this.API_URL}/auth/me`)
      );
      console.log('[AuthService] User loaded:', user?.email);
      this.userSignal.set(user);
    } catch (error) {
      console.log('[AuthService] Failed to load user, clearing tokens');
      this.clearTokens();
    } finally {
      this.loadingSignal.set(false);
    }
  }

  login(credentials: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.API_URL}/auth/login`, credentials).pipe(
      tap(response => {
        this.setTokens(response.accessToken, response.refreshToken);
        this.userSignal.set(response.user);
      })
    );
  }

  register(data: RegisterRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.API_URL}/auth/register`, data).pipe(
      tap(response => {
        this.setTokens(response.accessToken, response.refreshToken);
        this.userSignal.set(response.user);
      })
    );
  }

  refreshToken(): Observable<AuthResponse | null> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return of(null);
    }

    return this.http.post<AuthResponse>(`${this.API_URL}/auth/refresh`, { refreshToken }).pipe(
      tap(response => {
        this.setTokens(response.accessToken, response.refreshToken);
        this.userSignal.set(response.user);
      }),
      catchError(() => {
        this.logout();
        return of(null);
      })
    );
  }

  logout(): void {
    this.clearTokens();
    this.userSignal.set(null);
    this.router.navigate(['/auth/login']);
  }

  getAccessToken(): string | null {
    return localStorage.getItem('accessToken');
  }

  private getRefreshToken(): string | null {
    return localStorage.getItem('refreshToken');
  }

  private setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
  }

  private clearTokens(): void {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  }
}
