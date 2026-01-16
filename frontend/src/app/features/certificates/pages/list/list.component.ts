import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { SkeletonModule } from 'primeng/skeleton';
import { CertificateService, Certificate } from '@core/services/certificate.service';

@Component({
  selector: 'app-certificate-list',
  standalone: true,
  imports: [CommonModule, RouterLink, CardModule, ButtonModule, SkeletonModule],
  template: `
    <div class="page-container">
      <header class="page-header">
        <div>
          <h1>My Certificates</h1>
          <p>View and download your earned certificates</p>
        </div>
        <a routerLink="/dashboard">
          <p-button label="Back to Dashboard" icon="pi pi-arrow-left" [outlined]="true"></p-button>
        </a>
      </header>

      @if (loading()) {
        <div class="grid">
          @for (i of [1,2]; track i) {
            <div class="col-12 md:col-6">
              <p-skeleton height="200px"></p-skeleton>
            </div>
          }
        </div>
      } @else if (certificates().length === 0) {
        <div class="empty-state">
          <i class="pi pi-verified"></i>
          <h3>No certificates yet</h3>
          <p>Complete courses and pass tests to earn certificates</p>
          <a routerLink="/courses">
            <p-button label="Browse Courses" icon="pi pi-book"></p-button>
          </a>
        </div>
      } @else {
        <div class="grid">
          @for (cert of certificates(); track cert.id) {
            <div class="col-12 md:col-6">
              <div class="certificate-card">
                <div class="cert-header">
                  <div class="cert-icon">
                    <i class="pi pi-verified"></i>
                  </div>
                  <div class="cert-badge">Verified</div>
                </div>

                <h2>{{ cert.courseTitle }}</h2>

                <div class="cert-details">
                  <div class="detail">
                    <span class="label">Certificate Number</span>
                    <span class="value">{{ cert.certificateNumber }}</span>
                  </div>
                  <div class="detail">
                    <span class="label">Issued</span>
                    <span class="value">{{ cert.issuedAt | date:'mediumDate' }}</span>
                  </div>
                  <div class="detail">
                    <span class="label">Score</span>
                    <span class="value">{{ cert.score }}/{{ cert.maxScore }}</span>
                  </div>
                </div>

                <div class="cert-actions">
                  <p-button
                    label="Download PDF"
                    icon="pi pi-download"
                    (onClick)="downloadCertificate(cert)"
                  ></p-button>
                  <p-button
                    label="Share"
                    icon="pi pi-share-alt"
                    [outlined]="true"
                    (onClick)="shareCertificate(cert)"
                  ></p-button>
                </div>

                <div class="verify-link">
                  <span>Verification link:</span>
                  <code>{{ getVerifyUrl(cert) }}</code>
                </div>
              </div>
            </div>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .page-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
      flex-wrap: wrap;
      gap: 1rem;
    }

    .page-header h1 {
      margin: 0;
    }

    .page-header p {
      margin: 0.5rem 0 0;
      color: var(--text-secondary);
    }

    .empty-state {
      text-align: center;
      padding: 4rem 2rem;
    }

    .empty-state i {
      font-size: 4rem;
      color: var(--text-secondary);
    }

    .empty-state h3 {
      margin: 1rem 0 0.5rem;
    }

    .empty-state p {
      color: var(--text-secondary);
      margin-bottom: 1.5rem;
    }

    .certificate-card {
      background: linear-gradient(135deg, #1e3a5f 0%, #2d5a87 50%, #1e3a5f 100%);
      color: white;
      border-radius: 16px;
      padding: 2rem;
      position: relative;
      overflow: hidden;
    }

    .certificate-card::before {
      content: '';
      position: absolute;
      top: -50%;
      right: -50%;
      width: 100%;
      height: 100%;
      background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
    }

    .cert-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1.5rem;
    }

    .cert-icon {
      width: 50px;
      height: 50px;
      background: rgba(255, 255, 255, 0.2);
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .cert-icon i {
      font-size: 1.5rem;
    }

    .cert-badge {
      background: #22c55e;
      padding: 0.25rem 0.75rem;
      border-radius: 999px;
      font-size: 0.75rem;
      font-weight: 600;
    }

    .certificate-card h2 {
      margin: 0 0 1.5rem;
      font-size: 1.25rem;
    }

    .cert-details {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 1rem;
      margin-bottom: 1.5rem;
    }

    .detail {
      display: flex;
      flex-direction: column;
    }

    .detail .label {
      font-size: 0.75rem;
      opacity: 0.7;
      margin-bottom: 0.25rem;
    }

    .detail .value {
      font-weight: 600;
    }

    .cert-actions {
      display: flex;
      gap: 0.75rem;
      margin-bottom: 1.5rem;
    }

    .verify-link {
      background: rgba(0, 0, 0, 0.2);
      padding: 0.75rem;
      border-radius: 8px;
      font-size: 0.75rem;
    }

    .verify-link span {
      display: block;
      opacity: 0.7;
      margin-bottom: 0.25rem;
    }

    .verify-link code {
      word-break: break-all;
    }
  `]
})
export class CertificateListComponent implements OnInit {
  certificates = signal<Certificate[]>([]);
  loading = signal(true);

  constructor(private certificateService: CertificateService) {}

  ngOnInit(): void {
    this.loadCertificates();
  }

  loadCertificates(): void {
    this.certificateService.getMyCertificates().subscribe({
      next: (certificates) => {
        this.certificates.set(certificates);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  downloadCertificate(cert: Certificate): void {
    this.certificateService.downloadCertificate(cert.id).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `certificate-${cert.certificateNumber}.pdf`;
        a.click();
        window.URL.revokeObjectURL(url);
      }
    });
  }

  shareCertificate(cert: Certificate): void {
    const url = this.getVerifyUrl(cert);
    if (navigator.share) {
      navigator.share({
        title: `Certificate: ${cert.courseTitle}`,
        text: `I earned a certificate for completing ${cert.courseTitle}!`,
        url: url
      });
    } else {
      navigator.clipboard.writeText(url);
    }
  }

  getVerifyUrl(cert: Certificate): string {
    return `${window.location.origin}/certificates/verify/${cert.verificationHash}`;
  }
}
