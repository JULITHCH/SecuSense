import { Component, OnInit, Input, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { CertificateService, CertificateVerification } from '@core/services/certificate.service';

@Component({
  selector: 'app-verify',
  standalone: true,
  imports: [CommonModule, RouterLink, CardModule, ButtonModule],
  template: `
    <div class="verify-container">
      @if (loading()) {
        <div class="loading-state">
          <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
          <p>Verifying certificate...</p>
        </div>
      } @else if (verification()) {
        @if (verification()?.valid) {
          <div class="verification-card valid">
            <div class="status-icon">
              <i class="pi pi-verified"></i>
            </div>
            <h1>Certificate Verified</h1>
            <p class="status-message">This certificate is authentic and valid.</p>

            <div class="cert-details">
              <div class="detail-row">
                <span class="label">Certificate Number</span>
                <span class="value">{{ verification()?.certificateNumber }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Holder Name</span>
                <span class="value">{{ verification()?.holderName }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Course</span>
                <span class="value">{{ verification()?.courseTitle }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Issue Date</span>
                <span class="value">{{ verification()?.issuedAt | date:'mediumDate' }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Score</span>
                <span class="value">{{ verification()?.score }}/{{ verification()?.maxScore }}</span>
              </div>
            </div>
          </div>
        } @else {
          <div class="verification-card invalid">
            <div class="status-icon">
              <i class="pi pi-times-circle"></i>
            </div>
            <h1>Certificate Not Found</h1>
            <p class="status-message">This certificate could not be verified. It may be invalid or expired.</p>
          </div>
        }

        <div class="footer-actions">
          <a routerLink="/courses">
            <p-button label="Browse Courses" icon="pi pi-book" [outlined]="true"></p-button>
          </a>
        </div>
      }
    </div>
  `,
  styles: [`
    .verify-container {
      max-width: 500px;
      margin: 0 auto;
      padding: 2rem;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      justify-content: center;
    }

    .loading-state {
      text-align: center;
    }

    .verification-card {
      border-radius: 16px;
      padding: 3rem 2rem;
      text-align: center;
    }

    .verification-card.valid {
      background: linear-gradient(135deg, #10B981 0%, #059669 100%);
      color: white;
    }

    .verification-card.invalid {
      background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
      color: white;
    }

    .status-icon {
      margin-bottom: 1.5rem;
    }

    .status-icon i {
      font-size: 4rem;
    }

    .verification-card h1 {
      margin: 0 0 0.5rem;
    }

    .status-message {
      opacity: 0.9;
      margin-bottom: 2rem;
    }

    .cert-details {
      background: rgba(255, 255, 255, 0.1);
      border-radius: 12px;
      padding: 1.5rem;
      text-align: left;
    }

    .detail-row {
      display: flex;
      justify-content: space-between;
      padding: 0.75rem 0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }

    .detail-row:last-child {
      border-bottom: none;
    }

    .detail-row .label {
      opacity: 0.8;
    }

    .detail-row .value {
      font-weight: 600;
    }

    .footer-actions {
      text-align: center;
      margin-top: 2rem;
    }
  `]
})
export class VerifyComponent implements OnInit {
  @Input() hash!: string;

  verification = signal<CertificateVerification | null>(null);
  loading = signal(true);

  constructor(private certificateService: CertificateService) {}

  ngOnInit(): void {
    this.verifyCertificate();
  }

  verifyCertificate(): void {
    this.certificateService.verifyCertificate(this.hash).subscribe({
      next: (result) => {
        this.verification.set(result);
        this.loading.set(false);
      },
      error: () => {
        this.verification.set({ valid: false });
        this.loading.set(false);
      }
    });
  }
}
