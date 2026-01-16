import { Component, OnInit, Input, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { MessageService } from 'primeng/api';
import { TestService, TestResult } from '@core/services/test.service';
import { CertificateService } from '@core/services/certificate.service';

@Component({
  selector: 'app-results',
  standalone: true,
  imports: [CommonModule, RouterLink, CardModule, ButtonModule],
  template: `
    <div class="results-container">
      @if (loading()) {
        <div class="loading-state">
          <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
          <p>Loading results...</p>
        </div>
      } @else if (result()) {
        <div class="results-card" [class.passed]="result()?.passed" [class.failed]="!result()?.passed">
          <div class="result-icon">
            @if (result()?.passed) {
              <i class="pi pi-check-circle"></i>
            } @else {
              <i class="pi pi-times-circle"></i>
            }
          </div>

          <h1>{{ result()?.passed ? 'Congratulations!' : 'Keep Trying!' }}</h1>
          <p class="result-message">
            {{ result()?.passed
              ? 'You have successfully passed the test.'
              : 'You did not meet the passing score. Review the material and try again.' }}
          </p>

          <div class="score-display">
            <div class="score-circle">
              <span class="score-value">{{ result()?.percentage | number:'1.0-0' }}%</span>
              <span class="score-label">Score</span>
            </div>
          </div>

          <div class="score-details">
            <div class="detail-item">
              <span class="detail-value">{{ result()?.score }}</span>
              <span class="detail-label">Points Earned</span>
            </div>
            <div class="detail-item">
              <span class="detail-value">{{ result()?.maxScore }}</span>
              <span class="detail-label">Total Points</span>
            </div>
            <div class="detail-item">
              <span class="detail-value">{{ getCorrectCount() }}</span>
              <span class="detail-label">Correct Answers</span>
            </div>
          </div>

          @if (result()?.passed) {
            <div class="actions">
              @if (!certificateGenerated()) {
                <p-button
                  label="Get Certificate"
                  icon="pi pi-verified"
                  size="large"
                  [loading]="generatingCert()"
                  (onClick)="generateCertificate()"
                ></p-button>
              } @else {
                <p class="cert-success">
                  <i class="pi pi-check"></i> Certificate generated!
                </p>
                <a routerLink="/certificates">
                  <p-button label="View Certificates" icon="pi pi-eye" [outlined]="true"></p-button>
                </a>
              }
            </div>
          } @else {
            <div class="actions">
              <a [routerLink]="['/courses', courseId]">
                <p-button label="Review Course" icon="pi pi-book" [outlined]="true"></p-button>
              </a>
              <a [routerLink]="['/quiz', courseId]">
                <p-button label="Retry Test" icon="pi pi-refresh"></p-button>
              </a>
            </div>
          }

          <div class="answer-review">
            <h3>Answer Review</h3>
            @for (answer of result()?.answers; track answer.questionId; let i = $index) {
              <div class="answer-item" [class.correct]="answer.isCorrect" [class.incorrect]="!answer.isCorrect">
                <div class="answer-header">
                  <span class="question-num">Question {{ i + 1 }}</span>
                  <span class="answer-status">
                    @if (answer.isCorrect) {
                      <i class="pi pi-check"></i> Correct
                    } @else {
                      <i class="pi pi-times"></i> Incorrect
                    }
                  </span>
                  <span class="points">{{ answer.pointsAwarded }}/{{ answer.maxPoints }} pts</span>
                </div>
                @if (answer.explanation) {
                  <p class="explanation">{{ answer.explanation }}</p>
                }
              </div>
            }
          </div>

          <div class="back-link">
            <a routerLink="/dashboard">
              <p-button label="Back to Dashboard" icon="pi pi-arrow-left" [text]="true"></p-button>
            </a>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .results-container {
      max-width: 700px;
      margin: 0 auto;
      padding: 2rem;
    }

    .loading-state {
      text-align: center;
      padding: 4rem 2rem;
    }

    .results-card {
      background: var(--surface-card);
      border-radius: 16px;
      padding: 3rem 2rem;
      text-align: center;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    }

    .result-icon {
      margin-bottom: 1.5rem;
    }

    .result-icon i {
      font-size: 5rem;
    }

    .results-card.passed .result-icon i {
      color: var(--success-color);
    }

    .results-card.failed .result-icon i {
      color: var(--danger-color);
    }

    .results-card h1 {
      margin: 0 0 0.5rem;
    }

    .result-message {
      color: var(--text-secondary);
      margin-bottom: 2rem;
    }

    .score-display {
      margin: 2rem 0;
    }

    .score-circle {
      width: 140px;
      height: 140px;
      border-radius: 50%;
      background: var(--surface-ground);
      display: inline-flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
    }

    .results-card.passed .score-circle {
      background: linear-gradient(135deg, #10B981 0%, #059669 100%);
      color: white;
    }

    .results-card.failed .score-circle {
      background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
      color: white;
    }

    .score-value {
      font-size: 2.5rem;
      font-weight: 700;
    }

    .score-label {
      font-size: 0.875rem;
      opacity: 0.9;
    }

    .score-details {
      display: flex;
      justify-content: center;
      gap: 3rem;
      margin: 2rem 0;
      flex-wrap: wrap;
    }

    .detail-item {
      display: flex;
      flex-direction: column;
    }

    .detail-value {
      font-size: 1.5rem;
      font-weight: 600;
      color: var(--text-color);
    }

    .detail-label {
      font-size: 0.875rem;
      color: var(--text-secondary);
    }

    .actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
      margin: 2rem 0;
      flex-wrap: wrap;
    }

    .cert-success {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--success-color);
      font-weight: 500;
    }

    .answer-review {
      margin-top: 3rem;
      text-align: left;
    }

    .answer-review h3 {
      margin-bottom: 1rem;
    }

    .answer-item {
      padding: 1rem;
      border-radius: 8px;
      margin-bottom: 0.75rem;
    }

    .answer-item.correct {
      background: rgba(34, 197, 94, 0.1);
      border-left: 4px solid var(--success-color);
    }

    .answer-item.incorrect {
      background: rgba(239, 68, 68, 0.1);
      border-left: 4px solid var(--danger-color);
    }

    .answer-header {
      display: flex;
      align-items: center;
      gap: 1rem;
      flex-wrap: wrap;
    }

    .question-num {
      font-weight: 600;
    }

    .answer-status {
      display: flex;
      align-items: center;
      gap: 0.25rem;
      font-size: 0.875rem;
    }

    .answer-item.correct .answer-status {
      color: var(--success-color);
    }

    .answer-item.incorrect .answer-status {
      color: var(--danger-color);
    }

    .points {
      margin-left: auto;
      font-size: 0.875rem;
      color: var(--text-secondary);
    }

    .explanation {
      margin: 0.75rem 0 0;
      font-size: 0.875rem;
      color: var(--text-secondary);
      font-style: italic;
    }

    .back-link {
      margin-top: 2rem;
    }
  `]
})
export class ResultsComponent implements OnInit {
  @Input() attemptId!: string;

  result = signal<TestResult | null>(null);
  courseId: string = '';
  loading = signal(true);
  generatingCert = signal(false);
  certificateGenerated = signal(false);

  constructor(
    private testService: TestService,
    private certificateService: CertificateService,
    private messageService: MessageService,
    private router: Router
  ) {
    const nav = this.router.getCurrentNavigation();
    if (nav?.extras.state) {
      const state = nav.extras.state as any;
      if (state.result) {
        this.result.set(state.result);
        this.loading.set(false);
      }
      if (state.courseId) {
        this.courseId = state.courseId;
      }
    }
  }

  ngOnInit(): void {
    if (!this.result()) {
      this.loadResults();
    }
  }

  loadResults(): void {
    this.testService.getAttemptResults(this.attemptId).subscribe({
      next: (result) => {
        this.result.set(result);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  getCorrectCount(): number {
    return this.result()?.answers.filter(a => a.isCorrect).length || 0;
  }

  generateCertificate(): void {
    if (!this.courseId) {
      this.messageService.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Course information not available'
      });
      return;
    }

    this.generatingCert.set(true);
    this.certificateService.generateCertificate({
      courseId: this.courseId,
      attemptId: this.attemptId
    }).subscribe({
      next: () => {
        this.generatingCert.set(false);
        this.certificateGenerated.set(true);
        this.messageService.add({
          severity: 'success',
          summary: 'Certificate Generated!',
          detail: 'Your certificate is now available.'
        });
      },
      error: (err) => {
        this.generatingCert.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not generate certificate'
        });
      }
    });
  }
}
