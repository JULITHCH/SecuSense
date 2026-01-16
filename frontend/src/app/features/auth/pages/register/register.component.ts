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
  selector: 'app-register',
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
    <div class="register-container">
      <p-card header="Create Account" subheader="Join SecuSense Training Portal">
        <form [formGroup]="registerForm" (ngSubmit)="onSubmit()">
          <div class="grid">
            <div class="col-6">
              <div class="form-field">
                <label for="firstName">First Name</label>
                <input
                  pInputText
                  id="firstName"
                  formControlName="firstName"
                  placeholder="First name"
                />
                @if (registerForm.get('firstName')?.invalid && registerForm.get('firstName')?.touched) {
                  <small class="form-error">First name is required</small>
                }
              </div>
            </div>
            <div class="col-6">
              <div class="form-field">
                <label for="lastName">Last Name</label>
                <input
                  pInputText
                  id="lastName"
                  formControlName="lastName"
                  placeholder="Last name"
                />
                @if (registerForm.get('lastName')?.invalid && registerForm.get('lastName')?.touched) {
                  <small class="form-error">Last name is required</small>
                }
              </div>
            </div>
          </div>

          <div class="form-field">
            <label for="email">Email</label>
            <input
              pInputText
              id="email"
              formControlName="email"
              type="email"
              placeholder="Enter your email"
            />
            @if (registerForm.get('email')?.invalid && registerForm.get('email')?.touched) {
              <small class="form-error">Valid email is required</small>
            }
          </div>

          <div class="form-field">
            <label for="password">Password</label>
            <p-password
              id="password"
              formControlName="password"
              [toggleMask]="true"
              placeholder="Create a password"
              styleClass="w-full"
            ></p-password>
            @if (registerForm.get('password')?.invalid && registerForm.get('password')?.touched) {
              <small class="form-error">Password must be at least 8 characters</small>
            }
          </div>

          <p-button
            type="submit"
            label="Create Account"
            [loading]="loading"
            [disabled]="registerForm.invalid || loading"
            styleClass="w-full"
          ></p-button>
        </form>

        <ng-template pTemplate="footer">
          <div class="text-center mt-3">
            Already have an account?
            <a routerLink="/auth/login" class="text-primary">Sign in here</a>
          </div>
        </ng-template>
      </p-card>
    </div>
  `,
  styles: [`
    .register-container {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      padding: 1rem;
    }

    :host ::ng-deep .p-card {
      width: 100%;
      max-width: 500px;
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
export class RegisterComponent {
  registerForm: FormGroup;
  loading = false;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private messageService: MessageService
  ) {
    this.registerForm = this.fb.group({
      firstName: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(100)]],
      lastName: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(100)]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]]
    });
  }

  onSubmit(): void {
    if (this.registerForm.invalid) return;

    this.loading = true;
    this.authService.register(this.registerForm.value).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Welcome!',
          detail: 'Your account has been created successfully.'
        });
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.loading = false;
        this.messageService.add({
          severity: 'error',
          summary: 'Registration Failed',
          detail: err.error?.error || 'Could not create account'
        });
      }
    });
  }
}
