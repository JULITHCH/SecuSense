import { Component, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextareaModule } from 'primeng/inputtextarea';
import { DropdownModule } from 'primeng/dropdown';
import { SliderModule } from 'primeng/slider';
import { ButtonModule } from 'primeng/button';
import { ProgressBarModule } from 'primeng/progressbar';
import { MessageService } from 'primeng/api';
import { AIService, AIGenerationJob } from '@core/services/ai.service';

@Component({
  selector: 'app-course-generator',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    CardModule,
    InputTextModule,
    InputTextareaModule,
    DropdownModule,
    SliderModule,
    ButtonModule,
    ProgressBarModule
  ],
  template: `
    <div class="page-container">
      <header class="page-header">
        <div>
          <h1>AI Course Generator</h1>
          <p>Create courses automatically using AI</p>
        </div>
        <a routerLink="/admin">
          <p-button label="Back to Courses" icon="pi pi-arrow-left" [outlined]="true"></p-button>
        </a>
      </header>

      @if (!currentJob()) {
        <p-card header="Generate New Course">
          <form [formGroup]="generatorForm" (ngSubmit)="generateCourse()">
            <div class="form-field">
              <label for="topic">Course Topic *</label>
              <textarea
                pInputTextarea
                id="topic"
                formControlName="topic"
                rows="3"
                placeholder="Describe the topic you want to create a course about..."
              ></textarea>
              <small class="hint">Be specific about what you want the course to cover</small>
            </div>

            <div class="grid">
              <div class="col-12 md:col-6">
                <div class="form-field">
                  <label for="targetAudience">Target Audience</label>
                  <input
                    pInputText
                    id="targetAudience"
                    formControlName="targetAudience"
                    placeholder="e.g., IT professionals, beginners"
                  />
                </div>
              </div>

              <div class="col-12 md:col-6">
                <div class="form-field">
                  <label for="difficultyLevel">Difficulty Level</label>
                  <p-dropdown
                    id="difficultyLevel"
                    formControlName="difficultyLevel"
                    [options]="difficultyOptions"
                    placeholder="Select level"
                    styleClass="w-full"
                  ></p-dropdown>
                </div>
              </div>
            </div>

            <div class="form-field">
              <label>Video Duration: {{ generatorForm.get('videoDurationMin')?.value }} minutes</label>
              <p-slider
                formControlName="videoDurationMin"
                [min]="3"
                [max]="15"
              ></p-slider>
              <small class="hint">Recommended: 5-10 minutes for optimal engagement</small>
            </div>

            <div class="form-actions">
              <p-button
                type="submit"
                label="Generate Course"
                icon="pi pi-sparkles"
                [loading]="generating()"
                [disabled]="generatorForm.invalid || generating()"
              ></p-button>
            </div>
          </form>
        </p-card>
      } @else {
        <p-card header="Generation in Progress">
          <div class="generation-status">
            <div class="status-icon">
              @if (currentJob()?.status === 'completed') {
                <i class="pi pi-check-circle success"></i>
              } @else if (currentJob()?.status === 'failed') {
                <i class="pi pi-times-circle error"></i>
              } @else {
                <i class="pi pi-spin pi-spinner"></i>
              }
            </div>

            <h2>{{ getStatusTitle() }}</h2>
            <p class="status-message">{{ getStatusMessage() }}</p>

            @if (currentJob()?.status === 'processing' || currentJob()?.status === 'pending') {
              <p-progressBar mode="indeterminate"></p-progressBar>
            }

            @if (currentJob()?.status === 'completed') {
              <div class="video-status-info">
                <i class="pi pi-video"></i>
                <span>Video generation has been queued with Synthesia</span>
              </div>
              <div class="completion-actions">
                <a [routerLink]="['/admin']">
                  <p-button label="Manage Courses" icon="pi pi-list"></p-button>
                </a>
                <a [routerLink]="['/courses', currentJob()?.courseId]">
                  <p-button label="View Course" icon="pi pi-eye" [outlined]="true"></p-button>
                </a>
                <p-button
                  label="Generate Another"
                  icon="pi pi-plus"
                  [outlined]="true"
                  (onClick)="resetGenerator()"
                ></p-button>
              </div>
            }

            @if (currentJob()?.status === 'failed') {
              <div class="error-details">
                <p>{{ currentJob()?.error }}</p>
              </div>
              <p-button
                label="Try Again"
                icon="pi pi-refresh"
                (onClick)="resetGenerator()"
              ></p-button>
            }
          </div>
        </p-card>
      }

      <div class="info-section">
        <h3>How it works</h3>
        <div class="info-steps">
          <div class="step">
            <div class="step-number">1</div>
            <h4>Enter Topic</h4>
            <p>Describe what you want the course to cover</p>
          </div>
          <div class="step">
            <div class="step-number">2</div>
            <h4>AI Generation</h4>
            <p>AI creates content, video script, and quiz questions</p>
          </div>
          <div class="step">
            <div class="step-number">3</div>
            <h4>Review & Publish</h4>
            <p>Review the generated course and publish when ready</p>
          </div>
        </div>
      </div>
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

    .hint {
      color: var(--text-secondary);
      margin-top: 0.25rem;
      display: block;
    }

    .form-actions {
      margin-top: 2rem;
    }

    .generation-status {
      text-align: center;
      padding: 2rem;
    }

    .status-icon i {
      font-size: 4rem;
      margin-bottom: 1rem;
    }

    .status-icon .success {
      color: var(--success-color);
    }

    .status-icon .error {
      color: var(--danger-color);
    }

    .generation-status h2 {
      margin: 0 0 0.5rem;
    }

    .status-message {
      color: var(--text-secondary);
      margin-bottom: 2rem;
    }

    .video-status-info {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      padding: 1rem;
      background: rgba(59, 130, 246, 0.1);
      border-radius: 8px;
      color: var(--primary-color);
      margin-bottom: 1rem;
    }

    .completion-actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
      margin-top: 2rem;
      flex-wrap: wrap;
    }

    .error-details {
      background: rgba(239, 68, 68, 0.1);
      border-radius: 8px;
      padding: 1rem;
      margin: 1rem 0;
      color: var(--danger-color);
    }

    .info-section {
      margin-top: 3rem;
    }

    .info-section h3 {
      text-align: center;
      margin-bottom: 2rem;
    }

    .info-steps {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 2rem;
    }

    .step {
      text-align: center;
      padding: 1.5rem;
      background: var(--surface-ground);
      border-radius: 12px;
    }

    .step-number {
      width: 40px;
      height: 40px;
      background: var(--primary-color);
      color: white;
      border-radius: 50%;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-weight: 600;
      margin-bottom: 1rem;
    }

    .step h4 {
      margin: 0 0 0.5rem;
    }

    .step p {
      margin: 0;
      color: var(--text-secondary);
      font-size: 0.875rem;
    }
  `]
})
export class CourseGeneratorComponent implements OnDestroy {
  generatorForm: FormGroup;
  generating = signal(false);
  currentJob = signal<AIGenerationJob | null>(null);

  private pollInterval: any;

  difficultyOptions = [
    { label: 'Beginner', value: 'beginner' },
    { label: 'Intermediate', value: 'intermediate' },
    { label: 'Advanced', value: 'advanced' }
  ];

  constructor(
    private fb: FormBuilder,
    private aiService: AIService,
    private messageService: MessageService,
    private router: Router
  ) {
    this.generatorForm = this.fb.group({
      topic: ['', [Validators.required, Validators.minLength(5), Validators.maxLength(500)]],
      targetAudience: [''],
      difficultyLevel: ['intermediate'],
      videoDurationMin: [5]
    });
  }

  ngOnDestroy(): void {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
  }

  generateCourse(): void {
    if (this.generatorForm.invalid) return;

    this.generating.set(true);
    this.aiService.generateCourse(this.generatorForm.value).subscribe({
      next: (job) => {
        this.currentJob.set(job);
        this.generating.set(false);
        this.startPolling(job.id);
      },
      error: (err) => {
        this.generating.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Generation Failed',
          detail: err.error?.error || 'Could not start course generation'
        });
      }
    });
  }

  startPolling(jobId: string): void {
    this.pollInterval = setInterval(() => {
      this.aiService.getJob(jobId).subscribe({
        next: (job) => {
          this.currentJob.set(job);
          if (job.status === 'completed' || job.status === 'failed') {
            clearInterval(this.pollInterval);
            if (job.status === 'completed') {
              this.messageService.add({
                severity: 'success',
                summary: 'Course Generated!',
                detail: 'Your course is ready for review'
              });
            }
          }
        }
      });
    }, 3000);
  }

  resetGenerator(): void {
    this.currentJob.set(null);
    this.generatorForm.reset({
      difficultyLevel: 'intermediate',
      videoDurationMin: 5
    });
  }

  getStatusTitle(): string {
    const status = this.currentJob()?.status;
    switch (status) {
      case 'pending': return 'Queued';
      case 'processing': return 'Generating...';
      case 'completed': return 'Complete!';
      case 'failed': return 'Generation Failed';
      default: return 'Unknown';
    }
  }

  getStatusMessage(): string {
    const status = this.currentJob()?.status;
    switch (status) {
      case 'pending': return 'Your course is in the queue and will start generating soon.';
      case 'processing': return 'AI is creating your course content, video script, and quiz questions.';
      case 'completed': return 'Your course has been generated and is ready for review.';
      case 'failed': return 'Something went wrong during generation.';
      default: return '';
    }
  }
}
