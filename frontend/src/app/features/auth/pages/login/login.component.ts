import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { InputTextModule } from 'primeng/inputtext';
import { PasswordModule } from 'primeng/password';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { MessageService } from 'primeng/api';
import { AuthService } from '@core/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    InputTextModule,
    PasswordModule,
    ButtonModule,
    CardModule
  ],
  template: `
    <div class="login-container">
      <p-card header="Welcome to SecuSense" subheader="Sign in to continue">
        <form [formGroup]="loginForm" (ngSubmit)="onSubmit()">
          <div class="form-field">
            <label for="email">Email</label>
            <input
              pInputText
              id="email"
              formControlName="email"
              type="email"
              placeholder="Enter your email"
            />
            @if (loginForm.get('email')?.invalid && loginForm.get('email')?.touched) {
              <small class="form-error">Valid email is required</small>
            }
          </div>

          <div class="form-field">
            <label for="password">Password</label>
            <p-password
              id="password"
              formControlName="password"
              [feedback]="false"
              [toggleMask]="true"
              placeholder="Enter your password"
              styleClass="w-full"
            ></p-password>
            @if (loginForm.get('password')?.invalid && loginForm.get('password')?.touched) {
              <small class="form-error">Password is required</small>
            }
          </div>

          <p-button
            type="submit"
            label="Sign In"
            [loading]="loading"
            [disabled]="loginForm.invalid || loading"
            styleClass="w-full"
          ></p-button>
        </form>

        <ng-template pTemplate="footer">
          <div class="text-center mt-3">
            Don't have an account?
            <a routerLink="/auth/register" class="text-primary">Register here</a>
          </div>
        </ng-template>
      </p-card>
    </div>
  `,
  styles: [`
    .login-container {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      padding: 1rem;
    }

    :host ::ng-deep .p-card {
      width: 100%;
      max-width: 400px;
    }

    :host ::ng-deep .p-card-title {
      text-align: center;
      font-size: 1.5rem;
    }

    :host ::ng-deep .p-card-subtitle {
      text-align: center;
    }

    .text-primary {
      color: var(--primary-color);
      text-decoration: none;
    }

    .text-primary:hover {
      text-decoration: underline;
    }
  `]
})
export class LoginComponent {
  loginForm: FormGroup;
  loading = false;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private messageService: MessageService
  ) {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required]
    });
  }

  onSubmit(): void {
    if (this.loginForm.invalid) return;

    this.loading = true;
    this.authService.login(this.loginForm.value).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Welcome back!',
          detail: 'You have successfully logged in.'
        });
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.loading = false;
        this.messageService.add({
          severity: 'error',
          summary: 'Login Failed',
          detail: err.error?.error || 'Invalid credentials'
        });
      }
    });
  }
}
