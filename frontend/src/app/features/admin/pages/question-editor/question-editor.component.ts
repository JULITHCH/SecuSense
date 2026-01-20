import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextareaModule } from 'primeng/inputtextarea';
import { DropdownModule } from 'primeng/dropdown';
import { InputNumberModule } from 'primeng/inputnumber';
import { DialogModule } from 'primeng/dialog';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ToastModule } from 'primeng/toast';
import { MessageService, ConfirmationService } from 'primeng/api';
import { QuestionService, Question, Test, QuestionType } from '@core/services/question.service';

interface QuestionTypeOption {
  label: string;
  value: QuestionType;
}

@Component({
  selector: 'app-question-editor',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    InputTextareaModule,
    DropdownModule,
    InputNumberModule,
    DialogModule,
    ConfirmDialogModule,
    ToastModule
  ],
  providers: [MessageService, ConfirmationService],
  template: `
    <p-toast></p-toast>
    <p-confirmDialog></p-confirmDialog>

    <div class="p-4">
      <div class="flex align-items-center gap-3 mb-4">
        <p-button icon="pi pi-arrow-left" [text]="true" (onClick)="goBack()"></p-button>
        <h2 class="m-0">Edit Questions</h2>
      </div>

      @if (loading()) {
        <div class="text-center p-6">
          <i class="pi pi-spin pi-spinner text-4xl"></i>
          <p>Loading questions...</p>
        </div>
      } @else if (error()) {
        <p-card>
          <div class="text-center p-4">
            <i class="pi pi-exclamation-triangle text-4xl text-red-500 mb-3"></i>
            <p class="text-red-500">{{ error() }}</p>
            <p-button label="Go Back" icon="pi pi-arrow-left" (onClick)="goBack()"></p-button>
          </div>
        </p-card>
      } @else if (test()) {
        <p-card [header]="test()!.title" subheader="Test ID: {{ test()!.id }}">
          <div class="mb-3">
            <p>Passing Score: {{ test()!.passingScore }}%</p>
            <p>Total Questions: {{ questions().length }}</p>
          </div>

          <div class="flex justify-content-end mb-3">
            <p-button label="Add Question" icon="pi pi-plus" (onClick)="showAddDialog()"></p-button>
          </div>

          <div class="question-list">
            @for (question of questions(); track question.id; let i = $index) {
              <div class="question-item p-3 mb-3 border-1 surface-border border-round">
                <div class="flex justify-content-between align-items-start">
                  <div class="flex-grow-1">
                    <div class="flex align-items-center gap-2 mb-2">
                      <span class="font-bold">Q{{ i + 1 }}.</span>
                      <span class="p-tag p-tag-info">{{ formatQuestionType(question.questionType) }}</span>
                      <span class="p-tag p-tag-success">{{ question.points }} pts</span>
                    </div>
                    <p class="m-0">{{ question.questionText }}</p>
                  </div>
                  <div class="flex gap-2">
                    <p-button icon="pi pi-pencil" [text]="true" severity="info" (onClick)="editQuestion(question)"></p-button>
                    <p-button icon="pi pi-trash" [text]="true" severity="danger" (onClick)="confirmDelete(question)"></p-button>
                  </div>
                </div>
              </div>
            }

            @if (questions().length === 0) {
              <div class="text-center p-4 text-500">
                <i class="pi pi-question-circle text-4xl mb-3 block"></i>
                <p>No questions yet. Click "Add Question" to create one.</p>
              </div>
            }
          </div>
        </p-card>
      }
    </div>

    <!-- Add/Edit Question Dialog -->
    <p-dialog
      [(visible)]="showDialog"
      [header]="editingQuestion ? 'Edit Question' : 'Add Question'"
      [modal]="true"
      [style]="{width: '600px'}"
      [closable]="true">
      <div class="flex flex-column gap-3">
        <div>
          <label for="questionType" class="block mb-2">Question Type</label>
          <p-dropdown
            id="questionType"
            [options]="questionTypes"
            [(ngModel)]="dialogQuestion.questionType"
            optionLabel="label"
            optionValue="value"
            styleClass="w-full"
            [disabled]="editingQuestion !== null">
          </p-dropdown>
        </div>

        <div>
          <label for="questionText" class="block mb-2">Question Text</label>
          <textarea
            pInputTextarea
            id="questionText"
            [(ngModel)]="dialogQuestion.questionText"
            rows="3"
            class="w-full">
          </textarea>
        </div>

        <div>
          <label for="points" class="block mb-2">Points</label>
          <p-inputNumber
            id="points"
            [(ngModel)]="dialogQuestion.points"
            [min]="1"
            [max]="100">
          </p-inputNumber>
        </div>

        <!-- Question Data (JSON) -->
        <div>
          <label for="questionData" class="block mb-2">Question Data (JSON)</label>
          <textarea
            pInputTextarea
            id="questionData"
            [(ngModel)]="questionDataJson"
            rows="8"
            class="w-full font-mono text-sm">
          </textarea>
          <small class="text-500">
            @if (dialogQuestion.questionType === 'multiple_choice') {
              Format: {{'{'}}options: [...], correctIndices: [0], explanation: "..."{{'}'}}
            } @else if (dialogQuestion.questionType === 'drag_drop') {
              Format: {{'{'}}items: [...], dropZones: [...], correctMapping: {{'{}'}}, explanation: "..."{{'}'}}
            } @else if (dialogQuestion.questionType === 'fill_blank') {
              Format: {{'{'}}template: "Use {{'{{blank}}'}} here", blanks: [...], explanation: "..."{{'}'}}
            } @else if (dialogQuestion.questionType === 'matching') {
              Format: {{'{'}}leftItems: [...], rightItems: [...], correctPairs: {{'{}'}}, explanation: "..."{{'}'}}
            } @else if (dialogQuestion.questionType === 'ordering') {
              Format: {{'{'}}items: [...], correctOrder: [0,1,2], explanation: "..."{{'}'}}
            }
          </small>
        </div>
      </div>

      <ng-template pTemplate="footer">
        <p-button label="Cancel" [text]="true" (onClick)="hideDialog()"></p-button>
        <p-button [label]="editingQuestion ? 'Update' : 'Create'" icon="pi pi-check" (onClick)="saveQuestion()" [loading]="saving()"></p-button>
      </ng-template>
    </p-dialog>
  `,
  styles: [`
    .question-item {
      background: var(--surface-card);
      transition: box-shadow 0.2s;
    }
    .question-item:hover {
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
  `]
})
export class QuestionEditorComponent implements OnInit {
  courseId = '';
  test = signal<Test | null>(null);
  questions = signal<Question[]>([]);
  loading = signal(true);
  error = signal<string | null>(null);
  saving = signal(false);

  showDialog = false;
  editingQuestion: Question | null = null;
  dialogQuestion = {
    questionType: 'multiple_choice' as QuestionType,
    questionText: '',
    points: 10
  };
  questionDataJson = '';

  questionTypes: QuestionTypeOption[] = [
    { label: 'Multiple Choice', value: 'multiple_choice' },
    { label: 'Drag & Drop', value: 'drag_drop' },
    { label: 'Fill in the Blank', value: 'fill_blank' },
    { label: 'Matching', value: 'matching' },
    { label: 'Ordering', value: 'ordering' }
  ];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private questionService: QuestionService,
    private messageService: MessageService,
    private confirmationService: ConfirmationService
  ) {}

  ngOnInit(): void {
    this.courseId = this.route.snapshot.paramMap.get('courseId') || '';
    if (this.courseId) {
      this.loadQuestions();
    }
  }

  loadQuestions(): void {
    this.loading.set(true);
    this.error.set(null);

    this.questionService.getTestWithQuestions(this.courseId).subscribe({
      next: (test) => {
        this.test.set(test);
        this.questions.set(test.questions || []);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.error?.message || 'Failed to load questions. This course may not have a test yet.');
        this.loading.set(false);
      }
    });
  }

  formatQuestionType(type: QuestionType): string {
    const labels: Record<QuestionType, string> = {
      'multiple_choice': 'Multiple Choice',
      'drag_drop': 'Drag & Drop',
      'fill_blank': 'Fill Blank',
      'matching': 'Matching',
      'ordering': 'Ordering'
    };
    return labels[type] || type;
  }

  showAddDialog(): void {
    this.editingQuestion = null;
    this.dialogQuestion = {
      questionType: 'multiple_choice',
      questionText: '',
      points: 10
    };
    this.questionDataJson = JSON.stringify({
      options: ['Option A', 'Option B', 'Option C', 'Option D'],
      correctIndices: [0],
      explanation: 'Explanation here'
    }, null, 2);
    this.showDialog = true;
  }

  editQuestion(question: Question): void {
    this.editingQuestion = question;
    this.dialogQuestion = {
      questionType: question.questionType,
      questionText: question.questionText,
      points: question.points
    };
    this.questionDataJson = JSON.stringify(question.questionData, null, 2);
    this.showDialog = true;
  }

  hideDialog(): void {
    this.showDialog = false;
    this.editingQuestion = null;
  }

  saveQuestion(): void {
    let questionData: any;
    try {
      questionData = JSON.parse(this.questionDataJson);
    } catch (e) {
      this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Invalid JSON in question data' });
      return;
    }

    this.saving.set(true);

    if (this.editingQuestion) {
      // Update existing
      this.questionService.updateQuestion(this.editingQuestion.id, {
        questionType: this.dialogQuestion.questionType,
        questionText: this.dialogQuestion.questionText,
        questionData,
        points: this.dialogQuestion.points
      }).subscribe({
        next: (updated) => {
          this.questions.update(qs => qs.map(q => q.id === updated.id ? updated : q));
          this.messageService.add({ severity: 'success', summary: 'Success', detail: 'Question updated' });
          this.hideDialog();
          this.saving.set(false);
        },
        error: (err) => {
          this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to update question' });
          this.saving.set(false);
        }
      });
    } else {
      // Create new
      this.questionService.createQuestion({
        testId: this.test()!.id,
        questionType: this.dialogQuestion.questionType,
        questionText: this.dialogQuestion.questionText,
        questionData,
        points: this.dialogQuestion.points,
        orderIndex: this.questions().length
      }).subscribe({
        next: (created) => {
          this.questions.update(qs => [...qs, created]);
          this.messageService.add({ severity: 'success', summary: 'Success', detail: 'Question created' });
          this.hideDialog();
          this.saving.set(false);
        },
        error: (err) => {
          this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to create question' });
          this.saving.set(false);
        }
      });
    }
  }

  confirmDelete(question: Question): void {
    this.confirmationService.confirm({
      message: 'Are you sure you want to delete this question?',
      header: 'Delete Question',
      icon: 'pi pi-exclamation-triangle',
      accept: () => this.deleteQuestion(question)
    });
  }

  deleteQuestion(question: Question): void {
    this.questionService.deleteQuestion(question.id).subscribe({
      next: () => {
        this.questions.update(qs => qs.filter(q => q.id !== question.id));
        this.messageService.add({ severity: 'success', summary: 'Success', detail: 'Question deleted' });
      },
      error: (err) => {
        this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error?.message || 'Failed to delete question' });
      }
    });
  }

  goBack(): void {
    this.router.navigate(['/admin/courses']);
  }
}
