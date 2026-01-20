import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextareaModule } from 'primeng/inputtextarea';
import { DialogModule } from 'primeng/dialog';
import { ToastModule } from 'primeng/toast';
import { TagModule } from 'primeng/tag';
import { MessageService } from 'primeng/api';
import { environment } from '@env/environment';

interface LessonScript {
  id: string;
  sessionId: string;
  topicId: string;
  title: string;
  script: string;
  durationMin: number;
  outputType: 'video' | 'presentation';
  videoId?: string;
  videoUrl?: string;
  videoStatus?: string;
  presentationStatus?: string;
  sortOrder: number;
  createdAt: string;
}

interface CourseLessonWithPresentation {
  id: string;
  title: string;
  outputType: 'video' | 'presentation';
  videoUrl?: string;
  videoStatus?: string;
  presentationStatus?: string;
  presentation?: any;
}

@Component({
  selector: 'app-lesson-editor',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    InputTextareaModule,
    DialogModule,
    ToastModule,
    TagModule
  ],
  providers: [MessageService],
  template: `
    <p-toast></p-toast>

    <div class="p-4">
      <div class="flex align-items-center gap-3 mb-4">
        <p-button icon="pi pi-arrow-left" [text]="true" (onClick)="goBack()"></p-button>
        <h2 class="m-0">Edit Lessons</h2>
      </div>

      @if (loading()) {
        <div class="text-center p-6">
          <i class="pi pi-spin pi-spinner text-4xl"></i>
          <p>Loading lessons...</p>
        </div>
      } @else if (error()) {
        <p-card>
          <div class="text-center p-4">
            <i class="pi pi-exclamation-triangle text-4xl text-red-500 mb-3"></i>
            <p class="text-red-500">{{ error() }}</p>
            <p-button label="Go Back" icon="pi pi-arrow-left" (onClick)="goBack()"></p-button>
          </div>
        </p-card>
      } @else {
        <p-card header="Course Lessons" subheader="Edit lesson scripts and regenerate presentations">
          <div class="mb-3">
            <p>Total Lessons: {{ lessons().length }}</p>
          </div>

          <div class="lesson-list">
            @for (lesson of lessons(); track lesson.id; let i = $index) {
              <div class="lesson-item p-3 mb-3 border-1 surface-border border-round">
                <div class="flex justify-content-between align-items-start mb-3">
                  <div>
                    <h4 class="m-0 mb-2">{{ i + 1 }}. {{ lesson.title }}</h4>
                    <div class="flex gap-2">
                      <p-tag
                        [value]="lesson.outputType === 'video' ? 'Video' : 'Presentation'"
                        [severity]="lesson.outputType === 'video' ? 'info' : 'success'"
                        [icon]="lesson.outputType === 'video' ? 'pi pi-video' : 'pi pi-desktop'">
                      </p-tag>
                      @if (lesson.outputType === 'presentation') {
                        <p-tag
                          [value]="lesson.presentationStatus || 'Not Generated'"
                          [severity]="getStatusSeverity(lesson.presentationStatus)">
                        </p-tag>
                      }
                      @if (lesson.outputType === 'video' && lesson.videoUrl) {
                        <p-tag value="Video Ready" severity="success" icon="pi pi-check"></p-tag>
                      }
                    </div>
                  </div>
                  <div class="flex gap-2">
                    <p-button
                      icon="pi pi-pencil"
                      label="Edit Script"
                      [outlined]="true"
                      (onClick)="editLesson(lesson)">
                    </p-button>
                    <p-button
                      icon="pi pi-refresh"
                      label="Regenerate Script"
                      [outlined]="true"
                      severity="warning"
                      [loading]="regeneratingScript() === lesson.id"
                      (onClick)="regenerateScript(lesson)">
                    </p-button>
                    @if (lesson.outputType === 'presentation') {
                      <p-button
                        icon="pi pi-desktop"
                        label="Regenerate Presentation"
                        [outlined]="true"
                        severity="info"
                        [loading]="regeneratingPresentation() === lesson.id"
                        (onClick)="regeneratePresentation(lesson)">
                      </p-button>
                    }
                  </div>
                </div>
              </div>
            }

            @if (lessons().length === 0) {
              <div class="text-center p-4 text-500">
                <i class="pi pi-book text-4xl mb-3 block"></i>
                <p>No lessons found for this course.</p>
              </div>
            }
          </div>
        </p-card>
      }
    </div>

    <!-- Edit Script Dialog -->
    <p-dialog
      [(visible)]="showEditDialog"
      header="Edit Lesson Script"
      [modal]="true"
      [style]="{width: '800px'}"
      [closable]="true">
      @if (editingLesson) {
        <div class="flex flex-column gap-3">
          <div>
            <label for="lessonTitle" class="block mb-2 font-semibold">Title</label>
            <input
              pInputText
              id="lessonTitle"
              [(ngModel)]="editingLesson.title"
              class="w-full" />
          </div>

          <div>
            <label for="lessonScript" class="block mb-2 font-semibold">Script</label>
            <textarea
              pInputTextarea
              id="lessonScript"
              [(ngModel)]="editingLesson.script"
              rows="15"
              class="w-full font-mono text-sm">
            </textarea>
            <small class="text-500">This script is used for video narration or presentation audio.</small>
          </div>
        </div>
      }

      <ng-template pTemplate="footer">
        <p-button label="Cancel" [text]="true" (onClick)="hideDialog()"></p-button>
        <p-button label="Save" icon="pi pi-check" (onClick)="saveLesson()" [loading]="saving()"></p-button>
      </ng-template>
    </p-dialog>
  `,
  styles: [`
    .lesson-item {
      background: var(--surface-card);
      transition: box-shadow 0.2s;
    }
    .lesson-item:hover {
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
    .font-mono {
      font-family: 'Fira Code', 'Consolas', monospace;
    }
  `]
})
export class LessonEditorComponent implements OnInit {
  private readonly API_URL = environment.apiUrl;

  courseId = '';
  lessons = signal<CourseLessonWithPresentation[]>([]);
  loading = signal(true);
  error = signal<string | null>(null);
  saving = signal(false);
  regeneratingScript = signal<string | null>(null);
  regeneratingPresentation = signal<string | null>(null);

  showEditDialog = false;
  editingLesson: LessonScript | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private http: HttpClient,
    private messageService: MessageService
  ) {}

  ngOnInit(): void {
    this.courseId = this.route.snapshot.paramMap.get('courseId') || '';
    if (this.courseId) {
      this.loadLessons();
    }
  }

  loadLessons(): void {
    this.loading.set(true);
    this.error.set(null);

    this.http.get<CourseLessonWithPresentation[]>(`${this.API_URL}/courses/${this.courseId}/lessons`).subscribe({
      next: (lessons) => {
        this.lessons.set(lessons);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.error?.message || 'Failed to load lessons. This course may not have been created via the workflow.');
        this.loading.set(false);
      }
    });
  }

  getStatusSeverity(status?: string): 'success' | 'info' | 'warning' | 'danger' | 'secondary' {
    switch (status) {
      case 'completed': return 'success';
      case 'processing': return 'info';
      case 'failed': return 'danger';
      default: return 'secondary';
    }
  }

  editLesson(lesson: CourseLessonWithPresentation): void {
    // Load the full lesson script data
    this.http.get<LessonScript>(`${this.API_URL}/admin/courses/${this.courseId}/lessons/${lesson.id}`).subscribe({
      next: (fullLesson) => {
        this.editingLesson = { ...fullLesson };
        this.showEditDialog = true;
      },
      error: () => {
        // If we can't load from the new endpoint, use available data
        this.editingLesson = {
          id: lesson.id,
          sessionId: '',
          topicId: '',
          title: lesson.title,
          script: '',
          durationMin: 5,
          outputType: lesson.outputType,
          videoUrl: lesson.videoUrl,
          videoStatus: lesson.videoStatus,
          presentationStatus: lesson.presentationStatus,
          sortOrder: 0,
          createdAt: ''
        };
        this.showEditDialog = true;
        this.messageService.add({
          severity: 'warn',
          summary: 'Warning',
          detail: 'Could not load full lesson data. Some fields may be empty.'
        });
      }
    });
  }

  hideDialog(): void {
    this.showEditDialog = false;
    this.editingLesson = null;
  }

  saveLesson(): void {
    if (!this.editingLesson) return;

    this.saving.set(true);

    this.http.put(`${this.API_URL}/admin/courses/${this.courseId}/lessons/${this.editingLesson.id}`, {
      title: this.editingLesson.title,
      script: this.editingLesson.script
    }).subscribe({
      next: () => {
        // Update the lesson in the list
        this.lessons.update(ls => ls.map(l =>
          l.id === this.editingLesson!.id ? { ...l, title: this.editingLesson!.title } : l
        ));
        this.messageService.add({ severity: 'success', summary: 'Success', detail: 'Lesson script updated' });
        this.hideDialog();
        this.saving.set(false);
      },
      error: (err) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to update lesson' });
        this.saving.set(false);
      }
    });
  }

  regenerateScript(lesson: CourseLessonWithPresentation): void {
    this.regeneratingScript.set(lesson.id);

    this.http.post(`${this.API_URL}/admin/courses/${this.courseId}/lessons/${lesson.id}/regenerate`, {}).subscribe({
      next: (updated: any) => {
        this.lessons.update(ls => ls.map(l =>
          l.id === lesson.id ? { ...l, title: updated.title } : l
        ));
        this.messageService.add({ severity: 'success', summary: 'Success', detail: 'Lesson script regenerated' });
        this.regeneratingScript.set(null);
      },
      error: (err) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to regenerate script' });
        this.regeneratingScript.set(null);
      }
    });
  }

  regeneratePresentation(lesson: CourseLessonWithPresentation): void {
    this.regeneratingPresentation.set(lesson.id);

    this.http.post(`${this.API_URL}/admin/courses/${this.courseId}/lessons/${lesson.id}/regenerate-presentation`, {}).subscribe({
      next: () => {
        this.lessons.update(ls => ls.map(l =>
          l.id === lesson.id ? { ...l, presentationStatus: 'processing' } : l
        ));
        this.messageService.add({
          severity: 'success',
          summary: 'Started',
          detail: 'Presentation regeneration started. Refresh to check status.'
        });
        this.regeneratingPresentation.set(null);
      },
      error: (err) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to regenerate presentation' });
        this.regeneratingPresentation.set(null);
      }
    });
  }

  goBack(): void {
    this.router.navigate(['/admin/courses']);
  }
}
