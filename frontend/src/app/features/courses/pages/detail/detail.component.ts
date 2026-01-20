import { Component, OnInit, signal, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { ProgressBarModule } from 'primeng/progressbar';
import { SkeletonModule } from 'primeng/skeleton';
import { DialogModule } from 'primeng/dialog';
import { AccordionModule } from 'primeng/accordion';
import { MessageService } from 'primeng/api';
import { CourseService, Course, CourseLesson } from '@core/services/course.service';
import { EnrollmentService, Enrollment } from '@core/services/enrollment.service';
import { AuthService } from '@core/services/auth.service';
import { PresentationPlayerComponent } from '@shared/components/presentation-player/presentation-player.component';

@Component({
  selector: 'app-course-detail',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CardModule,
    ButtonModule,
    ProgressBarModule,
    SkeletonModule,
    DialogModule,
    AccordionModule,
    PresentationPlayerComponent
  ],
  template: `
    <div class="page-container">
      <a routerLink="/courses" class="back-link">
        <i class="pi pi-arrow-left"></i> Back to Courses
      </a>

      @if (loading()) {
        <div class="course-detail">
          <p-skeleton height="400px" styleClass="mb-3"></p-skeleton>
          <p-skeleton height="40px" styleClass="mb-2"></p-skeleton>
          <p-skeleton height="100px"></p-skeleton>
        </div>
      } @else if (course()) {
        <div class="course-detail">
          <!-- Lessons Section (for workflow-generated courses) -->
          @if (lessons().length > 0) {
            <div class="lessons-section">
              <h2 class="lessons-title">Course Content</h2>
              <div class="lessons-list">
                @for (lesson of lessons(); track lesson.id; let i = $index) {
                  <div class="lesson-card">
                    <div class="lesson-header">
                      <span class="lesson-number">{{ i + 1 }}</span>
                      <h3>{{ lesson.title }}</h3>
                      @if (lesson.outputType === 'presentation' && lesson.presentationStatus === 'completed') {
                        <span class="lesson-badge presentation">
                          <i class="pi pi-desktop"></i> Presentation
                        </span>
                      } @else if (lesson.outputType === 'video' && lesson.videoStatus === 'completed') {
                        <span class="lesson-badge video">
                          <i class="pi pi-video"></i> Video
                        </span>
                      }
                    </div>
                    <div class="lesson-actions">
                      @if (lesson.outputType === 'presentation' && lesson.presentation) {
                        <p-button
                          label="Start Lesson"
                          icon="pi pi-play"
                          (onClick)="openPresentation(lesson)"
                        ></p-button>
                      } @else if (lesson.outputType === 'video' && lesson.videoUrl) {
                        <p-button
                          label="Watch Video"
                          icon="pi pi-play"
                          (onClick)="playVideo(lesson.videoUrl!)"
                        ></p-button>
                      } @else {
                        <span class="coming-soon">Coming soon</span>
                      }
                    </div>
                  </div>
                }
              </div>
            </div>
          } @else {
            <!-- Fallback to single video (for quick-generated courses) -->
            <div class="video-section">
              @if (course()?.videoUrl) {
                <div class="video-container">
                  <video
                    #videoPlayer
                    [src]="course()?.videoUrl"
                    controls
                    (timeupdate)="onVideoProgress($event)"
                    (ended)="onVideoEnded()"
                  >
                    Your browser does not support the video tag.
                  </video>
                </div>
              } @else {
                <div class="video-placeholder">
                  @if (course()?.videoStatus === 'pending' || course()?.videoStatus === 'processing') {
                    <i class="pi pi-spin pi-spinner"></i>
                    <p>Video is being generated...</p>
                    <small>This may take a few minutes</small>
                  } @else if (course()?.videoStatus === 'failed') {
                    <i class="pi pi-exclamation-circle"></i>
                    <p>Video generation failed</p>
                    <small>{{ course()?.videoError }}</small>
                  } @else {
                    <i class="pi pi-video"></i>
                    <p>Video content coming soon</p>
                  }
                </div>
              }
            </div>
          }

          <div class="course-info">
            <h1>{{ course()?.title }}</h1>
            <p class="description">{{ course()?.description }}</p>

            <div class="course-meta">
              <span><i class="pi pi-check-circle"></i> Pass: {{ course()?.passPercentage }}%</span>
            </div>

            @if (authService.isAuthenticated()) {
              @if (enrollment()) {
                <div class="enrollment-status">
                  <h3>Your Progress</h3>
                  <p-progressBar [value]="enrollment()?.progressPercentage || 0"></p-progressBar>
                  <p class="progress-text">{{ enrollment()?.progressPercentage }}% completed</p>

                  @if (enrollment()?.videoWatched || !course()?.videoUrl) {
                    <p-button
                      label="Take the Test"
                      icon="pi pi-pencil"
                      (onClick)="startTest()"
                      styleClass="mt-3"
                    ></p-button>
                  } @else {
                    <p class="video-notice">
                      <i class="pi pi-info-circle"></i>
                      Watch the full video to unlock the test
                    </p>
                  }
                </div>
              } @else {
                <p-button
                  label="Enroll in Course"
                  icon="pi pi-plus"
                  [loading]="enrolling()"
                  (onClick)="enroll()"
                  size="large"
                ></p-button>
              }
            } @else {
              <div class="login-prompt">
                <p>Sign in to enroll in this course</p>
                <a routerLink="/auth/login">
                  <p-button label="Sign In" icon="pi pi-sign-in"></p-button>
                </a>
              </div>
            }
          </div>
        </div>
      } @else {
        <div class="not-found">
          <h2>Course not found</h2>
          <a routerLink="/courses">
            <p-button label="Back to Courses" icon="pi pi-arrow-left"></p-button>
          </a>
        </div>
      }

      <!-- Presentation Dialog -->
      <p-dialog
        header="Lesson"
        [(visible)]="showPresentationDialog"
        [modal]="true"
        [style]="{ width: '90vw', height: '90vh' }"
        [maximizable]="true"
        (onHide)="onPresentationClose()"
      >
        @if (currentLesson()?.presentation) {
          <app-presentation-player
            [slides]="currentLesson()!.presentation!.slides"
            [autoPlay]="true">
          </app-presentation-player>
        }
      </p-dialog>

      <!-- Video Dialog -->
      <p-dialog
        header="Video"
        [(visible)]="showVideoDialog"
        [modal]="true"
        [style]="{ width: '90vw', height: '90vh' }"
        [maximizable]="true"
        (onHide)="onVideoDialogClose()"
      >
        @if (currentVideoUrl()) {
          <div class="video-dialog-container">
            <video
              #dialogVideoPlayer
              [src]="currentVideoUrl()"
              controls
              autoplay
            >
              Your browser does not support the video tag.
            </video>
          </div>
        }
      </p-dialog>
    </div>
  `,
  styles: [`
    .back-link {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--text-secondary);
      text-decoration: none;
      margin-bottom: 1.5rem;
    }

    .back-link:hover {
      color: var(--primary-color);
    }

    .course-detail {
      display: grid;
      gap: 2rem;
    }

    .video-container {
      position: relative;
      width: 100%;
      padding-top: 56.25%;
      background: #000;
      border-radius: 12px;
      overflow: hidden;
    }

    .video-container video {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
    }

    .video-placeholder {
      width: 100%;
      height: 400px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border-radius: 12px;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: white;
    }

    .video-placeholder i {
      font-size: 4rem;
      margin-bottom: 1rem;
    }

    .course-info h1 {
      margin: 0 0 1rem;
      color: var(--text-color);
    }

    .description {
      color: var(--text-secondary);
      line-height: 1.6;
      margin-bottom: 1.5rem;
    }

    .course-meta {
      display: flex;
      gap: 1.5rem;
      margin-bottom: 2rem;
      color: var(--text-secondary);
    }

    .course-meta span {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .enrollment-status {
      background: var(--surface-ground);
      border-radius: 12px;
      padding: 1.5rem;
    }

    .enrollment-status h3 {
      margin: 0 0 1rem;
    }

    .progress-text {
      margin-top: 0.5rem;
      color: var(--text-secondary);
    }

    .video-notice {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: 1rem;
      padding: 0.75rem 1rem;
      background: #fef3c7;
      border-radius: 8px;
      color: #92400e;
    }

    .login-prompt {
      background: var(--surface-ground);
      border-radius: 12px;
      padding: 2rem;
      text-align: center;
    }

    .login-prompt p {
      margin: 0 0 1rem;
      color: var(--text-secondary);
    }

    .not-found {
      text-align: center;
      padding: 4rem 2rem;
    }

    .lessons-section {
      margin-bottom: 2rem;
    }

    .lessons-title {
      margin: 0 0 1.5rem;
      color: var(--text-color);
    }

    .lessons-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .lesson-card {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1.25rem 1.5rem;
      background: var(--surface-card);
      border-radius: 12px;
      border: 1px solid var(--surface-border);
      transition: box-shadow 0.2s;
    }

    .lesson-card:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    .lesson-header {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .lesson-number {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: var(--primary-color);
      color: white;
      font-weight: 600;
      font-size: 0.875rem;
    }

    .lesson-header h3 {
      margin: 0;
      font-size: 1rem;
      font-weight: 500;
    }

    .lesson-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.25rem 0.75rem;
      border-radius: 16px;
      font-size: 0.75rem;
      font-weight: 500;
    }

    .lesson-badge.presentation {
      background: #e0f2fe;
      color: #0369a1;
    }

    .lesson-badge.video {
      background: #fef3c7;
      color: #92400e;
    }

    .lesson-actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .coming-soon {
      color: var(--text-secondary);
      font-size: 0.875rem;
      font-style: italic;
    }

    .video-dialog-container {
      width: 100%;
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #000;
    }

    .video-dialog-container video {
      max-width: 100%;
      max-height: 100%;
    }
  `]
})
export class CourseDetailComponent implements OnInit {
  @Input() id!: string;

  course = signal<Course | null>(null);
  lessons = signal<CourseLesson[]>([]);
  enrollment = signal<Enrollment | null>(null);
  loading = signal(true);
  enrolling = signal(false);

  // Presentation dialog
  showPresentationDialog = false;
  currentLesson = signal<CourseLesson | null>(null);

  // Video dialog
  showVideoDialog = false;
  currentVideoUrl = signal<string | null>(null);

  constructor(
    private courseService: CourseService,
    private enrollmentService: EnrollmentService,
    public authService: AuthService,
    private messageService: MessageService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadCourse();
    this.loadLessons();
    if (this.authService.isAuthenticated()) {
      this.loadEnrollment();
    }
  }

  loadCourse(): void {
    this.courseService.getCourseById(this.id).subscribe({
      next: (course) => {
        this.course.set(course);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  loadLessons(): void {
    this.courseService.getCourseLessons(this.id).subscribe({
      next: (lessons) => {
        this.lessons.set(lessons);
      },
      error: () => {
        // No lessons available (e.g., quick-generated course)
        this.lessons.set([]);
      }
    });
  }

  openPresentation(lesson: CourseLesson): void {
    this.currentLesson.set(lesson);
    this.showPresentationDialog = true;
  }

  onPresentationClose(): void {
    this.currentLesson.set(null);
  }

  playVideo(url: string): void {
    this.currentVideoUrl.set(url);
    this.showVideoDialog = true;
  }

  onVideoDialogClose(): void {
    this.currentVideoUrl.set(null);
  }

  loadEnrollment(): void {
    this.enrollmentService.getMyEnrollments().subscribe({
      next: (enrollments) => {
        const enrollment = enrollments.find(e => e.courseId === this.id);
        if (enrollment) {
          this.enrollment.set(enrollment);
        }
      }
    });
  }

  enroll(): void {
    this.enrolling.set(true);
    this.enrollmentService.enrollInCourse(this.id).subscribe({
      next: (enrollment) => {
        this.enrollment.set(enrollment);
        this.enrolling.set(false);
        this.messageService.add({
          severity: 'success',
          summary: 'Enrolled!',
          detail: 'You have successfully enrolled in this course.'
        });
      },
      error: (err) => {
        this.enrolling.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Enrollment Failed',
          detail: err.error?.error || 'Could not enroll in course'
        });
      }
    });
  }

  onVideoProgress(event: Event): void {
    const video = event.target as HTMLVideoElement;
    const progress = Math.round((video.currentTime / video.duration) * 100);

    const enrollment = this.enrollment();
    if (enrollment && progress > (enrollment.progressPercentage || 0)) {
      this.enrollmentService.updateProgress(enrollment.id, progress).subscribe({
        next: (updated) => this.enrollment.set(updated)
      });
    }
  }

  onVideoEnded(): void {
    const enrollment = this.enrollment();
    if (enrollment && !enrollment.videoWatched) {
      this.enrollmentService.completeVideo(enrollment.id).subscribe({
        next: (updated) => {
          this.enrollment.set(updated);
          this.messageService.add({
            severity: 'success',
            summary: 'Video Completed!',
            detail: 'You can now take the test.'
          });
        }
      });
    }
  }

  startTest(): void {
    this.router.navigate(['/quiz', this.id]);
  }
}
