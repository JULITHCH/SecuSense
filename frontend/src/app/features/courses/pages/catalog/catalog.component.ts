import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { SkeletonModule } from 'primeng/skeleton';
import { PaginatorModule, PaginatorState } from 'primeng/paginator';
import { CourseService, Course } from '@core/services/course.service';
import { AuthService } from '@core/services/auth.service';

@Component({
  selector: 'app-catalog',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CardModule,
    ButtonModule,
    SkeletonModule,
    PaginatorModule
  ],
  template: `
    <div class="page-container">
      <header class="catalog-header">
        <div class="header-content">
          <h1>Training Courses</h1>
          <p>Enhance your security knowledge with our comprehensive courses</p>
        </div>
        @if (authService.isAuthenticated()) {
          <a routerLink="/dashboard">
            <p-button label="My Dashboard" icon="pi pi-user" [outlined]="true"></p-button>
          </a>
        } @else {
          <a routerLink="/auth/login">
            <p-button label="Sign In" icon="pi pi-sign-in"></p-button>
          </a>
        }
      </header>

      @if (loading()) {
        <div class="grid">
          @for (i of [1,2,3,4,5,6]; track i) {
            <div class="col-12 md:col-6 lg:col-4">
              <p-card>
                <p-skeleton height="150px"></p-skeleton>
                <p-skeleton styleClass="mt-2" height="24px"></p-skeleton>
                <p-skeleton styleClass="mt-2" height="60px"></p-skeleton>
              </p-card>
            </div>
          }
        </div>
      } @else {
        @if (courses().length === 0) {
          <div class="empty-state">
            <i class="pi pi-book" style="font-size: 4rem; color: var(--text-secondary)"></i>
            <h3>No courses available</h3>
            <p>Check back later for new training content.</p>
          </div>
        } @else {
          <div class="grid">
            @for (course of courses(); track course.id) {
              <div class="col-12 md:col-6 lg:col-4">
                <p-card [header]="course.title" styleClass="course-card">
                  @if (course.thumbnailUrl) {
                    <img [src]="course.thumbnailUrl" [alt]="course.title" class="course-thumbnail" />
                  } @else {
                    <div class="course-thumbnail-placeholder">
                      <i class="pi pi-video"></i>
                    </div>
                  }
                  <p class="course-description">{{ course.description | slice:0:120 }}{{ course.description.length > 120 ? '...' : '' }}</p>
                  <ng-template pTemplate="footer">
                    <a [routerLink]="['/courses', course.id]">
                      <p-button label="View Course" icon="pi pi-arrow-right" iconPos="right"></p-button>
                    </a>
                  </ng-template>
                </p-card>
              </div>
            }
          </div>

          @if (totalRecords() > pageSize) {
            <p-paginator
              [rows]="pageSize"
              [totalRecords]="totalRecords()"
              [first]="(currentPage() - 1) * pageSize"
              (onPageChange)="onPageChange($event)"
            ></p-paginator>
          }
        }
      }
    </div>
  `,
  styles: [`
    .catalog-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
      flex-wrap: wrap;
      gap: 1rem;
    }

    .header-content h1 {
      margin: 0;
      color: var(--text-color);
    }

    .header-content p {
      margin: 0.5rem 0 0;
      color: var(--text-secondary);
    }

    .course-card {
      height: 100%;
    }

    .course-thumbnail {
      width: 100%;
      height: 150px;
      object-fit: cover;
      border-radius: 6px;
      margin-bottom: 1rem;
    }

    .course-thumbnail-placeholder {
      width: 100%;
      height: 150px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border-radius: 6px;
      margin-bottom: 1rem;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .course-thumbnail-placeholder i {
      font-size: 3rem;
      color: white;
    }

    .course-description {
      color: var(--text-secondary);
      line-height: 1.5;
    }

    .empty-state {
      text-align: center;
      padding: 4rem 2rem;
    }

    .empty-state h3 {
      margin: 1rem 0 0.5rem;
      color: var(--text-color);
    }

    .empty-state p {
      color: var(--text-secondary);
    }
  `]
})
export class CatalogComponent implements OnInit {
  courses = signal<Course[]>([]);
  loading = signal(true);
  totalRecords = signal(0);
  currentPage = signal(1);
  pageSize = 9;

  constructor(
    private courseService: CourseService,
    public authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadCourses();
  }

  loadCourses(): void {
    this.loading.set(true);
    this.courseService.getCourses(this.currentPage(), this.pageSize).subscribe({
      next: (response) => {
        this.courses.set(response.data);
        this.totalRecords.set(response.total);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  onPageChange(event: PaginatorState): void {
    this.currentPage.set(Math.floor((event.first ?? 0) / this.pageSize) + 1);
    this.loadCourses();
  }
}
