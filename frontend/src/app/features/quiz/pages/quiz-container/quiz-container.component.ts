import { Component, OnInit, Input, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { RadioButtonModule } from 'primeng/radiobutton';
import { CheckboxModule } from 'primeng/checkbox';
import { InputTextModule } from 'primeng/inputtext';
import { DragDropModule } from 'primeng/dragdrop';
import { ProgressBarModule } from 'primeng/progressbar';
import { FormsModule } from '@angular/forms';
import { MessageService } from 'primeng/api';
import { TestService, Test, Question, SubmitAnswer } from '@core/services/test.service';

@Component({
  selector: 'app-quiz-container',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    FormsModule,
    CardModule,
    ButtonModule,
    RadioButtonModule,
    CheckboxModule,
    InputTextModule,
    DragDropModule,
    ProgressBarModule
  ],
  template: `
    <div class="quiz-container">
      @if (loading()) {
        <div class="loading-state">
          <i class="pi pi-spin pi-spinner" style="font-size: 2rem"></i>
          <p>Loading test...</p>
        </div>
      } @else if (!test()) {
        <div class="error-state">
          <i class="pi pi-exclamation-circle"></i>
          <h3>Test not available</h3>
          <p>This course doesn't have a test yet.</p>
          <a [routerLink]="['/courses', courseId]">
            <p-button label="Back to Course" icon="pi pi-arrow-left"></p-button>
          </a>
        </div>
      } @else if (!attemptId()) {
        <div class="start-screen">
          <h1>{{ test()?.title }}</h1>
          <p>{{ test()?.description }}</p>
          <div class="test-info">
            <div class="info-item">
              <i class="pi pi-question-circle"></i>
              <span>{{ test()?.questions?.length }} questions</span>
            </div>
            <div class="info-item">
              <i class="pi pi-check-circle"></i>
              <span>{{ test()?.passingScore }}% to pass</span>
            </div>
            @if (test()?.timeLimitMinutes) {
              <div class="info-item">
                <i class="pi pi-clock"></i>
                <span>{{ test()?.timeLimitMinutes }} minutes</span>
              </div>
            }
          </div>
          <p-button
            label="Start Test"
            icon="pi pi-play"
            size="large"
            [loading]="starting()"
            (onClick)="startTest()"
          ></p-button>
        </div>
      } @else {
        <div class="quiz-header">
          <h2>{{ test()?.title }}</h2>
          <div class="progress-info">
            Question {{ currentQuestionIndex() + 1 }} of {{ test()?.questions?.length }}
          </div>
          <p-progressBar [value]="progressPercent()"></p-progressBar>
        </div>

        @if (currentQuestion(); as question) {
          <div class="question-card">
            <div class="question-text">{{ question.questionText }}</div>

            @switch (question.questionType) {
              @case ('multiple_choice') {
                <div class="options-list">
                  @for (option of question.questionData.options; track $index) {
                    <div
                      class="option-item"
                      [class.selected]="isOptionSelected($index)"
                      (click)="toggleMultipleChoice($index)"
                    >
                      <span class="option-letter">{{ getOptionLetter($index) }}</span>
                      <span class="option-text">{{ option }}</span>
                      @if (isOptionSelected($index)) {
                        <i class="pi pi-check"></i>
                      }
                    </div>
                  }
                </div>
              }

              @case ('fill_blank') {
                <div class="fill-blank-section">
                  @for (blank of fillBlankAnswers(); track $index; let i = $index) {
                    <div class="blank-input">
                      <label>Blank {{ i + 1 }}</label>
                      <input
                        pInputText
                        [value]="blank"
                        (input)="updateFillBlank(i, $event)"
                        placeholder="Enter your answer"
                      />
                    </div>
                  }
                </div>
              }

              @case ('matching') {
                <div class="matching-section">
                  <div class="matching-columns">
                    <div class="left-column">
                      @for (item of question.questionData.leftItems; track item) {
                        <div class="match-item">
                          <span>{{ item }}</span>
                          <select (change)="updateMatching(item, $event)">
                            <option value="">Select match...</option>
                            @for (right of question.questionData.rightItems; track right) {
                              <option [value]="right" [selected]="matchingAnswers()[item] === right">{{ right }}</option>
                            }
                          </select>
                        </div>
                      }
                    </div>
                  </div>
                </div>
              }

              @case ('ordering') {
                <div class="ordering-section">
                  <p class="ordering-hint">Drag items to reorder them</p>
                  <div class="ordering-list">
                    @for (item of orderingItems(); track item; let i = $index) {
                      <div
                        class="order-item"
                        pDraggable="ordering"
                        pDroppable="ordering"
                        (onDragStart)="dragStart(i)"
                        (onDrop)="drop(i)"
                      >
                        <span class="order-number">{{ i + 1 }}</span>
                        <span>{{ item }}</span>
                        <i class="pi pi-bars drag-handle"></i>
                      </div>
                    }
                  </div>
                </div>
              }

              @case ('drag_drop') {
                <div class="drag-drop-section">
                  <div class="drag-items">
                    <h4>Items to place:</h4>
                    <div class="items-container">
                      @for (item of getUnplacedItems(); track item) {
                        <div
                          class="draggable-item"
                          pDraggable="dragdrop"
                          (onDragStart)="dragDropStart(item)"
                        >
                          {{ item }}
                        </div>
                      }
                    </div>
                  </div>
                  <div class="drop-zones">
                    <h4>Drop zones:</h4>
                    @for (zone of currentQuestion()?.questionData.dropZones; track zone) {
                      <div
                        class="drop-zone"
                        pDroppable="dragdrop"
                        (onDrop)="dragDropPlace(zone)"
                        [class.has-item]="dragDropAnswers()[zone]"
                      >
                        <span class="zone-label">{{ zone }}</span>
                        @if (dragDropAnswers()[zone]; as placedItem) {
                          <div class="placed-item">
                            {{ getItemForZone(zone) }}
                            <button class="remove-btn" (click)="removeDragDropItem(zone)">
                              <i class="pi pi-times"></i>
                            </button>
                          </div>
                        }
                      </div>
                    }
                  </div>
                </div>
              }
            }
          </div>
        }

        <div class="quiz-navigation">
          <p-button
            label="Previous"
            icon="pi pi-arrow-left"
            [outlined]="true"
            [disabled]="currentQuestionIndex() === 0"
            (onClick)="previousQuestion()"
          ></p-button>

          @if (currentQuestionIndex() < (test()?.questions?.length || 0) - 1) {
            <p-button
              label="Next"
              icon="pi pi-arrow-right"
              iconPos="right"
              (onClick)="nextQuestion()"
            ></p-button>
          } @else {
            <p-button
              label="Submit Test"
              icon="pi pi-check"
              severity="success"
              [loading]="submitting()"
              (onClick)="submitTest()"
            ></p-button>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .quiz-container {
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem;
    }

    .loading-state, .error-state {
      text-align: center;
      padding: 4rem 2rem;
    }

    .error-state i {
      font-size: 4rem;
      color: var(--warning-color);
    }

    .start-screen {
      text-align: center;
      padding: 2rem;
    }

    .start-screen h1 {
      margin-bottom: 1rem;
    }

    .test-info {
      display: flex;
      justify-content: center;
      gap: 2rem;
      margin: 2rem 0;
    }

    .info-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--text-secondary);
    }

    .quiz-header {
      margin-bottom: 2rem;
    }

    .quiz-header h2 {
      margin: 0 0 0.5rem;
    }

    .progress-info {
      color: var(--text-secondary);
      margin-bottom: 0.5rem;
    }

    .question-card {
      background: var(--surface-card);
      border-radius: 12px;
      padding: 2rem;
      margin-bottom: 2rem;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }

    .question-text {
      font-size: 1.125rem;
      font-weight: 500;
      margin-bottom: 1.5rem;
      line-height: 1.5;
    }

    .options-list {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }

    .option-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1rem;
      border: 2px solid #e2e8f0;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.2s ease;
    }

    .option-item:hover {
      border-color: var(--primary-color);
      background: rgba(59, 130, 246, 0.05);
    }

    .option-item.selected {
      border-color: var(--primary-color);
      background: rgba(59, 130, 246, 0.1);
    }

    .option-letter {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: var(--surface-ground);
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 600;
    }

    .option-item.selected .option-letter {
      background: var(--primary-color);
      color: white;
    }

    .option-text {
      flex: 1;
    }

    .fill-blank-section {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .blank-input {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .blank-input label {
      font-weight: 500;
      color: var(--text-secondary);
    }

    .matching-section .match-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1rem;
      padding: 0.75rem;
      background: var(--surface-ground);
      border-radius: 8px;
    }

    .match-item span {
      flex: 1;
    }

    .match-item select {
      padding: 0.5rem;
      border: 1px solid #e2e8f0;
      border-radius: 6px;
      min-width: 200px;
    }

    .ordering-section .ordering-hint {
      color: var(--text-secondary);
      margin-bottom: 1rem;
    }

    .ordering-list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .order-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1rem;
      background: var(--surface-ground);
      border-radius: 8px;
      cursor: grab;
    }

    .order-item:active {
      cursor: grabbing;
    }

    .order-number {
      width: 28px;
      height: 28px;
      background: var(--primary-color);
      color: white;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 600;
    }

    .drag-handle {
      margin-left: auto;
      color: var(--text-secondary);
    }

    .drag-drop-section {
      display: grid;
      gap: 2rem;
    }

    .drag-drop-section h4 {
      margin: 0 0 0.75rem;
    }

    .items-container {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .draggable-item {
      padding: 0.75rem 1rem;
      background: var(--surface-ground);
      border: 1px solid #e2e8f0;
      border-radius: 6px;
      cursor: grab;
    }

    .drop-zone {
      min-height: 60px;
      border: 2px dashed #cbd5e1;
      border-radius: 8px;
      padding: 1rem;
      margin-bottom: 0.75rem;
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .drop-zone.has-item {
      border-style: solid;
      border-color: var(--success-color);
      background: rgba(34, 197, 94, 0.05);
    }

    .zone-label {
      font-weight: 500;
      color: var(--text-secondary);
    }

    .placed-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.5rem;
      background: var(--surface-card);
      border-radius: 4px;
    }

    .remove-btn {
      background: none;
      border: none;
      cursor: pointer;
      color: var(--danger-color);
      padding: 0.25rem;
    }

    .quiz-navigation {
      display: flex;
      justify-content: space-between;
    }
  `]
})
export class QuizContainerComponent implements OnInit {
  @Input() courseId!: string;

  test = signal<Test | null>(null);
  attemptId = signal<string | null>(null);
  loading = signal(true);
  starting = signal(false);
  submitting = signal(false);

  currentQuestionIndex = signal(0);
  answers = signal<Map<string, any>>(new Map());

  // Question-specific state
  selectedOptions = signal<number[]>([]);
  fillBlankAnswers = signal<string[]>([]);
  matchingAnswers = signal<Record<string, string>>({});
  orderingItems = signal<string[]>([]);
  dragDropAnswers = signal<Record<string, string>>({});
  currentDragItem = signal<string | null>(null);
  currentDragIndex = signal<number>(-1);

  currentQuestion = computed(() => {
    const t = this.test();
    if (!t?.questions) return null;
    return t.questions[this.currentQuestionIndex()];
  });

  progressPercent = computed(() => {
    const t = this.test();
    if (!t?.questions) return 0;
    return Math.round(((this.currentQuestionIndex() + 1) / t.questions.length) * 100);
  });

  constructor(
    private testService: TestService,
    private router: Router,
    private messageService: MessageService
  ) {}

  ngOnInit(): void {
    this.loadTest();
  }

  loadTest(): void {
    this.testService.getTestByCourseId(this.courseId).subscribe({
      next: (test) => {
        this.test.set(test);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  startTest(): void {
    const t = this.test();
    if (!t) return;

    this.starting.set(true);
    this.testService.startAttempt(t.id).subscribe({
      next: (attempt) => {
        this.attemptId.set(attempt.id);
        this.starting.set(false);
        this.initializeQuestionState();
      },
      error: (err) => {
        this.starting.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not start test'
        });
      }
    });
  }

  initializeQuestionState(): void {
    const question = this.currentQuestion();
    if (!question) return;

    switch (question.questionType) {
      case 'multiple_choice':
        this.selectedOptions.set([]);
        break;
      case 'fill_blank':
        const blanksCount = question.questionData.template?.match(/\{\{blank\}\}/g)?.length || 0;
        this.fillBlankAnswers.set(new Array(blanksCount).fill(''));
        break;
      case 'matching':
        this.matchingAnswers.set({});
        break;
      case 'ordering':
        this.orderingItems.set([...question.questionData.items]);
        break;
      case 'drag_drop':
        this.dragDropAnswers.set({});
        break;
    }
  }

  saveCurrentAnswer(): void {
    const question = this.currentQuestion();
    if (!question) return;

    let answerData: any;
    switch (question.questionType) {
      case 'multiple_choice':
        answerData = this.selectedOptions();
        break;
      case 'fill_blank':
        answerData = this.fillBlankAnswers();
        break;
      case 'matching':
        answerData = this.matchingAnswers();
        break;
      case 'ordering':
        answerData = this.orderingItems().map((_, i) => i);
        break;
      case 'drag_drop':
        const mapping: Record<string, string> = {};
        Object.entries(this.dragDropAnswers()).forEach(([zone, item]) => {
          if (item) mapping[item] = zone;
        });
        answerData = mapping;
        break;
    }

    const currentAnswers = new Map(this.answers());
    currentAnswers.set(question.id, answerData);
    this.answers.set(currentAnswers);
  }

  loadSavedAnswer(): void {
    const question = this.currentQuestion();
    if (!question) return;

    const savedAnswer = this.answers().get(question.id);
    if (!savedAnswer) {
      this.initializeQuestionState();
      return;
    }

    switch (question.questionType) {
      case 'multiple_choice':
        this.selectedOptions.set(savedAnswer);
        break;
      case 'fill_blank':
        this.fillBlankAnswers.set(savedAnswer);
        break;
      case 'matching':
        this.matchingAnswers.set(savedAnswer);
        break;
      case 'ordering':
        this.orderingItems.set(question.questionData.items);
        break;
      case 'drag_drop':
        const zoneMapping: Record<string, string> = {};
        Object.entries(savedAnswer as Record<string, string>).forEach(([item, zone]) => {
          zoneMapping[zone] = item;
        });
        this.dragDropAnswers.set(zoneMapping);
        break;
    }
  }

  previousQuestion(): void {
    this.saveCurrentAnswer();
    this.currentQuestionIndex.update(i => i - 1);
    this.loadSavedAnswer();
  }

  nextQuestion(): void {
    this.saveCurrentAnswer();
    this.currentQuestionIndex.update(i => i + 1);
    this.loadSavedAnswer();
  }

  submitTest(): void {
    this.saveCurrentAnswer();

    const t = this.test();
    const attempt = this.attemptId();
    if (!t?.questions || !attempt) return;

    const submitAnswers: SubmitAnswer[] = t.questions.map(q => ({
      questionId: q.id,
      answerData: this.answers().get(q.id) || null
    }));

    this.submitting.set(true);
    this.testService.submitAttempt(attempt, submitAnswers).subscribe({
      next: (result) => {
        this.router.navigate(['/quiz/results', attempt], {
          state: { result, courseId: this.courseId }
        });
      },
      error: (err) => {
        this.submitting.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Submission Failed',
          detail: err.error?.error || 'Could not submit test'
        });
      }
    });
  }

  // Multiple choice helpers
  getOptionLetter(index: number): string {
    return String.fromCharCode(65 + index);
  }

  isOptionSelected(index: number): boolean {
    return this.selectedOptions().includes(index);
  }

  toggleMultipleChoice(index: number): void {
    const current = this.selectedOptions();
    if (current.includes(index)) {
      this.selectedOptions.set(current.filter(i => i !== index));
    } else {
      this.selectedOptions.set([...current, index]);
    }
  }

  // Fill blank helpers
  updateFillBlank(index: number, event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    const answers = [...this.fillBlankAnswers()];
    answers[index] = value;
    this.fillBlankAnswers.set(answers);
  }

  // Matching helpers
  updateMatching(leftItem: string, event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    const answers = { ...this.matchingAnswers() };
    if (value) {
      answers[leftItem] = value;
    } else {
      delete answers[leftItem];
    }
    this.matchingAnswers.set(answers);
  }

  // Ordering helpers
  dragStart(index: number): void {
    this.currentDragIndex.set(index);
  }

  drop(targetIndex: number): void {
    const sourceIndex = this.currentDragIndex();
    if (sourceIndex === -1 || sourceIndex === targetIndex) return;

    const items = [...this.orderingItems()];
    const [removed] = items.splice(sourceIndex, 1);
    items.splice(targetIndex, 0, removed);
    this.orderingItems.set(items);
    this.currentDragIndex.set(-1);
  }

  // Drag drop helpers
  dragDropStart(item: string): void {
    this.currentDragItem.set(item);
  }

  dragDropPlace(zone: string): void {
    const item = this.currentDragItem();
    if (!item) return;

    const answers = { ...this.dragDropAnswers() };
    // Remove item from any existing zone
    Object.keys(answers).forEach(z => {
      if (answers[z] === item) delete answers[z];
    });
    answers[zone] = item;
    this.dragDropAnswers.set(answers);
    this.currentDragItem.set(null);
  }

  removeDragDropItem(zone: string): void {
    const answers = { ...this.dragDropAnswers() };
    delete answers[zone];
    this.dragDropAnswers.set(answers);
  }

  getUnplacedItems(): string[] {
    const question = this.currentQuestion();
    if (!question) return [];
    const placedItems = Object.values(this.dragDropAnswers());
    return question.questionData.items.filter((item: string) => !placedItems.includes(item));
  }

  getItemForZone(zone: string): string {
    return this.dragDropAnswers()[zone] || '';
  }
}
