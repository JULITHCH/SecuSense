import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { ProgressBarModule } from 'primeng/progressbar';
import { TabViewModule } from 'primeng/tabview';
import { SkeletonModule } from 'primeng/skeleton';
import { AuthService } from '@core/services/auth.service';
import { EnrollmentService, EnrollmentWithCourse } from '@core/services/enrollment.service';
import { CertificateService, Certificate } from '@core/services/certificate.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CardModule,
    ButtonModule,
    ProgressBarModule,
    TabViewModule,
    SkeletonModule
  ],
  template: `
    <div class="page-container">
      <header class="dashboard-header">
        <div>
          <h1>Welcome, {{ authService.user()?.firstName }}!</h1>
          <p>Track your progress and certificates</p>
        </div>
        <div class="header-actions">
          <a routerLink="/courses">
            <p-button label="Browse Courses" icon="pi pi-book" [outlined]="true"></p-button>
          </a>
          @if (authService.isAdmin()) {
            <a routerLink="/admin">
              <p-button label="Admin Panel" icon="pi pi-cog"></p-button>
            </a>
          }
          <p-button label="Sign Out" icon="pi pi-sign-out" severity="secondary" (onClick)="logout()"></p-button>
        </div>
      </header>

      <p-tabView>
        <p-tabPanel header="My Courses">
          @if (loadingEnrollments()) {
            <div class="grid">
              @for (i of [1,2,3]; track i) {
                <div class="col-12 md:col-6 lg:col-4">
                  <p-skeleton height="200px"></p-skeleton>
                </div>
              }
            </div>
          } @else if (enrollments().length === 0) {
            <div class="empty-state">
              <i class="pi pi-book"></i>
              <h3>No courses yet</h3>
              <p>Start learning by enrolling in a course</p>
              <a routerLink="/courses">
                <p-button label="Browse Courses" icon="pi pi-arrow-right"></p-button>
              </a>
            </div>
          } @else {
            <div class="grid">
              @for (enrollment of enrollments(); track enrollment.id) {
                <div class="col-12 md:col-6 lg:col-4">
                  <p-card [header]="enrollment.courseTitle" styleClass="enrollment-card">
                    <div class="progress-section">
                      <p-progressBar [value]="enrollment.progressPercentage"></p-progressBar>
                      <span class="progress-text">{{ enrollment.progressPercentage }}% complete</span>
                    </div>
                    <div class="status-badges">
                      @if (enrollment.videoWatched) {
                        <span class="badge badge-success">
                          <i class="pi pi-check"></i> Video completed
                        </span>
                      }
                      @if (enrollment.status === 'completed') {
                        <span class="badge badge-primary">
                          <i class="pi pi-trophy"></i> Course completed
                        </span>
                      }
                    </div>
                    <ng-template pTemplate="footer">
                      <a [routerLink]="['/courses', enrollment.courseId]">
                        <p-button label="Continue" icon="pi pi-play"></p-button>
                      </a>
                      @if (enrollment.videoWatched) {
                        <a [routerLink]="['/quiz', enrollment.courseId]">
                          <p-button label="Take Test" icon="pi pi-pencil" [outlined]="true"></p-button>
                        </a>
                      }
                    </ng-template>
                  </p-card>
                </div>
              }
            </div>
          }
        </p-tabPanel>

        <p-tabPanel header="My Certificates">
          @if (loadingCertificates()) {
            <div class="grid">
              @for (i of [1,2]; track i) {
                <div class="col-12 md:col-6">
                  <p-skeleton height="150px"></p-skeleton>
                </div>
              }
            </div>
          } @else if (certificates().length === 0) {
            <div class="empty-state">
              <i class="pi pi-verified"></i>
              <h3>No certificates yet</h3>
              <p>Complete courses to earn certificates</p>
            </div>
          } @else {
            <div class="grid">
              @for (cert of certificates(); track cert.id) {
                <div class="col-12 md:col-6">
                  <div class="certificate-card">
                    <div class="cert-icon">
                      <i class="pi pi-verified"></i>
                    </div>
                    <div class="cert-info">
                      <h3>{{ cert.courseTitle }}</h3>
                      <p>Certificate #{{ cert.certificateNumber }}</p>
                      <p class="cert-date">Issued: {{ cert.issuedAt | date:'mediumDate' }}</p>
                    </div>
                    <div class="cert-actions">
                      <a [routerLink]="['/certificates', cert.id]">
                        <p-button icon="pi pi-eye" [rounded]="true" [text]="true" pTooltip="View"></p-button>
                      </a>
                      <p-button icon="pi pi-download" [rounded]="true" [text]="true" pTooltip="Download" (onClick)="downloadCertificate(cert)"></p-button>
                    </div>
                  </div>
                </div>
              }
            </div>
          }
        </p-tabPanel>
      </p-tabView>
    </div>
  `,
  styles: [`
    .dashboard-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
      flex-wrap: wrap;
      gap: 1rem;
    }

    .dashboard-header h1 {
      margin: 0;
    }

    .dashboard-header p {
      margin: 0.5rem 0 0;
      color: var(--text-secondary);
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
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

    .enrollment-card {
      height: 100%;
    }

    .progress-section {
      margin-bottom: 1rem;
    }

    .progress-text {
      font-size: 0.875rem;
      color: var(--text-secondary);
    }

    .status-badges {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }

    .badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.25rem 0.75rem;
      border-radius: 999px;
      font-size: 0.75rem;
      font-weight: 500;
    }

    .badge-success {
      background: #d1fae5;
      color: #065f46;
    }

    .badge-primary {
      background: #dbeafe;
      color: #1e40af;
    }

    .certificate-card {
      display: flex;
      align-items: center;
      gap: 1rem;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      border-radius: 12px;
      padding: 1.5rem;
    }

    .cert-icon {
      flex-shrink: 0;
      width: 60px;
      height: 60px;
      background: rgba(255, 255, 255, 0.2);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .cert-icon i {
      font-size: 1.5rem;
    }

    .cert-info {
      flex: 1;
    }

    .cert-info h3 {
      margin: 0 0 0.25rem;
    }

    .cert-info p {
      margin: 0;
      opacity: 0.9;
      font-size: 0.875rem;
    }

    .cert-date {
      opacity: 0.7 !important;
    }

    .cert-actions {
      display: flex;
      gap: 0.5rem;
    }

    .cert-actions :host ::ng-deep .p-button {
      color: white;
    }
  `]
})
export class DashboardComponent implements OnInit {
  enrollments = signal<EnrollmentWithCourse[]>([]);
  certificates = signal<Certificate[]>([]);
  loadingEnrollments = signal(true);
  loadingCertificates = signal(true);

  constructor(
    public authService: AuthService,
    private enrollmentService: EnrollmentService,
    private certificateService: CertificateService
  ) {}

  ngOnInit(): void {
    this.loadEnrollments();
    this.loadCertificates();
  }

  loadEnrollments(): void {
    this.enrollmentService.getMyEnrollments().subscribe({
      next: (enrollments) => {
        this.enrollments.set(enrollments);
        this.loadingEnrollments.set(false);
      },
      error: () => this.loadingEnrollments.set(false)
    });
  }

  loadCertificates(): void {
    this.certificateService.getMyCertificates().subscribe({
      next: (certificates) => {
        this.certificates.set(certificates);
        this.loadingCertificates.set(false);
      },
      error: () => this.loadingCertificates.set(false)
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

  logout(): void {
    this.authService.logout();
  }
}
