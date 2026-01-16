import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { TagModule } from 'primeng/tag';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ConfirmationService, MessageService } from 'primeng/api';
import { CourseService, Course } from '@core/services/course.service';

@Component({
  selector: 'app-admin-course-list',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    TableModule,
    ButtonModule,
    TagModule,
    ConfirmDialogModule
  ],
  providers: [ConfirmationService],
  template: `
    <div class="page-container">
      <p-confirmDialog></p-confirmDialog>

      <header class="page-header">
        <div>
          <h1>Course Management</h1>
          <p>Manage and generate training courses</p>
        </div>
        <div class="header-actions">
          <a routerLink="/dashboard">
            <p-button label="Dashboard" icon="pi pi-home" [outlined]="true"></p-button>
          </a>
          <a routerLink="/admin/generate">
            <p-button label="Generate Course" icon="pi pi-sparkles"></p-button>
          </a>
        </div>
      </header>

      <p-table
        [value]="courses()"
        [loading]="loading()"
        [paginator]="true"
        [rows]="10"
        [rowsPerPageOptions]="[10, 25, 50]"
        [totalRecords]="totalRecords()"
        [lazy]="true"
        (onLazyLoad)="loadCourses($event)"
        styleClass="p-datatable-striped"
      >
        <ng-template pTemplate="header">
          <tr>
            <th>Title</th>
            <th>Status</th>
            <th>Video</th>
            <th>Pass %</th>
            <th>Created</th>
            <th style="width: 200px">Actions</th>
          </tr>
        </ng-template>
        <ng-template pTemplate="body" let-course>
          <tr>
            <td>
              <div class="course-title">
                <strong>{{ course.title }}</strong>
                <small>{{ course.description | slice:0:50 }}...</small>
              </div>
            </td>
            <td>
              <p-tag
                [value]="course.isPublished ? 'Published' : 'Draft'"
                [severity]="course.isPublished ? 'success' : 'warning'"
              ></p-tag>
            </td>
            <td>
              @if (course.videoUrl) {
                <p-tag value="Ready" severity="success" icon="pi pi-check"></p-tag>
              } @else if (course.videoStatus === 'pending' || course.videoStatus === 'processing') {
                <p-tag value="Generating" severity="info" icon="pi pi-spin pi-spinner"></p-tag>
              } @else if (course.videoStatus === 'failed') {
                <p-tag value="Failed" severity="danger" icon="pi pi-times"></p-tag>
              } @else {
                <p-tag value="None" severity="secondary"></p-tag>
              }
            </td>
            <td>{{ course.passPercentage }}%</td>
            <td>{{ course.createdAt | date:'shortDate' }}</td>
            <td>
              <div class="action-buttons">
                <a [routerLink]="['/courses', course.id]">
                  <p-button icon="pi pi-eye" [rounded]="true" [text]="true" pTooltip="View"></p-button>
                </a>
                @if (course.isPublished) {
                  <p-button
                    icon="pi pi-eye-slash"
                    [rounded]="true"
                    [text]="true"
                    severity="warning"
                    pTooltip="Unpublish"
                    (onClick)="unpublishCourse(course)"
                  ></p-button>
                } @else {
                  <p-button
                    icon="pi pi-check"
                    [rounded]="true"
                    [text]="true"
                    severity="success"
                    pTooltip="Publish"
                    (onClick)="publishCourse(course)"
                  ></p-button>
                }
                <p-button
                  icon="pi pi-trash"
                  [rounded]="true"
                  [text]="true"
                  severity="danger"
                  pTooltip="Delete"
                  (onClick)="confirmDelete(course)"
                ></p-button>
              </div>
            </td>
          </tr>
        </ng-template>
        <ng-template pTemplate="emptymessage">
          <tr>
            <td colspan="6" class="text-center">
              <div class="empty-state-inline">
                <i class="pi pi-book"></i>
                <span>No courses found. Generate your first course!</span>
              </div>
            </td>
          </tr>
        </ng-template>
      </p-table>
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

    .header-actions {
      display: flex;
      gap: 0.5rem;
    }

    .course-title {
      display: flex;
      flex-direction: column;
    }

    .course-title small {
      color: var(--text-secondary);
    }

    .action-buttons {
      display: flex;
      gap: 0.25rem;
    }

    .empty-state-inline {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.75rem;
      padding: 2rem;
      color: var(--text-secondary);
    }

    .empty-state-inline i {
      font-size: 1.5rem;
    }
  `]
})
export class AdminCourseListComponent implements OnInit {
  courses = signal<Course[]>([]);
  loading = signal(true);
  totalRecords = signal(0);

  constructor(
    private courseService: CourseService,
    private confirmationService: ConfirmationService,
    private messageService: MessageService
  ) {}

  ngOnInit(): void {
    this.loadCourses({ first: 0, rows: 10 });
  }

  loadCourses(event: any): void {
    this.loading.set(true);
    const page = Math.floor(event.first / event.rows) + 1;

    this.courseService.getAllCourses(page, event.rows).subscribe({
      next: (response) => {
        this.courses.set(response.data);
        this.totalRecords.set(response.total);
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  publishCourse(course: Course): void {
    this.courseService.publishCourse(course.id).subscribe({
      next: () => {
        course.isPublished = true;
        this.messageService.add({
          severity: 'success',
          summary: 'Published',
          detail: 'Course is now visible to users'
        });
      },
      error: () => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Could not publish course'
        });
      }
    });
  }

  unpublishCourse(course: Course): void {
    this.courseService.unpublishCourse(course.id).subscribe({
      next: () => {
        course.isPublished = false;
        this.messageService.add({
          severity: 'info',
          summary: 'Unpublished',
          detail: 'Course is now hidden from users'
        });
      },
      error: () => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Could not unpublish course'
        });
      }
    });
  }

  confirmDelete(course: Course): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to delete "${course.title}"?`,
      header: 'Delete Course',
      icon: 'pi pi-exclamation-triangle',
      accept: () => this.deleteCourse(course)
    });
  }

  deleteCourse(course: Course): void {
    this.courseService.deleteCourse(course.id).subscribe({
      next: () => {
        this.courses.update(courses => courses.filter(c => c.id !== course.id));
        this.messageService.add({
          severity: 'success',
          summary: 'Deleted',
          detail: 'Course has been deleted'
        });
      },
      error: () => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Could not delete course'
        });
      }
    });
  }
}
