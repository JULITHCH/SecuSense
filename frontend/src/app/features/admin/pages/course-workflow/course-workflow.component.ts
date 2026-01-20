import { Component, signal, OnDestroy, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, FormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextareaModule } from 'primeng/inputtextarea';
import { DropdownModule } from 'primeng/dropdown';
import { SliderModule } from 'primeng/slider';
import { ButtonModule } from 'primeng/button';
import { ProgressBarModule } from 'primeng/progressbar';
import { StepsModule } from 'primeng/steps';
import { DialogModule } from 'primeng/dialog';
import { PanelModule } from 'primeng/panel';
import { ChipsModule } from 'primeng/chips';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { TooltipModule } from 'primeng/tooltip';
import { MenuItem, MessageService, ConfirmationService } from 'primeng/api';
import {
  WorkflowService,
  CourseWorkflowSession,
  TopicSuggestion,
  RefinedTopic,
  LessonScript,
  WorkflowStep,
  OutputType,
  LessonPresentation,
  UpdateRefinedTopicRequest,
  ReorderTopicsRequest,
  UpdateLessonScriptRequest
} from '@core/services/workflow.service';
import { PresentationPlayerComponent } from '@shared/components/presentation-player/presentation-player.component';

@Component({
  selector: 'app-course-workflow',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    FormsModule,
    RouterLink,
    CardModule,
    InputTextModule,
    InputTextareaModule,
    DropdownModule,
    SliderModule,
    ButtonModule,
    ProgressBarModule,
    StepsModule,
    DialogModule,
    PanelModule,
    ChipsModule,
    ConfirmDialogModule,
    TooltipModule,
    PresentationPlayerComponent
  ],
  providers: [ConfirmationService],
  template: `
    <div class="page-container">
      <header class="page-header">
        <div>
          <h1>AI Course Workflow</h1>
          <p>Create courses with AI-powered multi-step generation</p>
        </div>
        <a routerLink="/admin">
          <p-button label="Back to Courses" icon="pi pi-arrow-left" [outlined]="true"></p-button>
        </a>
      </header>

      <!-- Step Indicator -->
      <div class="workflow-steps">
        <p-steps [model]="workflowSteps" [activeIndex]="currentStepIndex()" [readonly]="true"></p-steps>
      </div>

      <!-- Step 1: Topic Entry -->
      @if (!session()) {
        <p-card header="Step 1: Enter Course Topic">
          <form [formGroup]="topicForm" (ngSubmit)="startResearch()">
            <div class="form-field">
              <label for="topic">Course Topic *</label>
              <textarea
                pInputTextarea
                id="topic"
                formControlName="topic"
                rows="3"
                placeholder="Describe the main topic for your course..."
              ></textarea>
              <small class="hint">The AI will research this topic and suggest specific lessons</small>
            </div>

            <div class="grid">
              <div class="col-12 md:col-4">
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

              <div class="col-12 md:col-4">
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

              <div class="col-12 md:col-4">
                <div class="form-field">
                  <label for="language">Course Language *</label>
                  <p-dropdown
                    id="language"
                    formControlName="language"
                    [options]="languageOptions"
                    placeholder="Select language"
                    styleClass="w-full"
                  ></p-dropdown>
                </div>
              </div>
            </div>

            <div class="form-field">
              <label>Video Duration: {{ topicForm.get('videoDurationMin')?.value }} minutes</label>
              <p-slider
                formControlName="videoDurationMin"
                [min]="3"
                [max]="15"
              ></p-slider>
              <small class="hint">Per-lesson video duration</small>
            </div>

            <div class="form-actions">
              <p-button
                type="submit"
                label="Start Research"
                icon="pi pi-search"
                [loading]="loading()"
                [disabled]="topicForm.invalid || loading()"
              ></p-button>
            </div>
          </form>
        </p-card>
      }

      <!-- Processing States -->
      @if (session() && isProcessing()) {
        <p-card [header]="getProcessingTitle()">
          <div class="processing-status">
            <div class="status-icon">
              <i class="pi pi-spin pi-spinner"></i>
            </div>
            <h2>{{ getProcessingMessage() }}</h2>
            <p-progressBar mode="indeterminate"></p-progressBar>
          </div>
        </p-card>
      }

      <!-- Error State -->
      @if (session() && session()!.status === 'failed') {
        <p-card header="Generation Failed">
          <div class="error-status">
            <div class="status-icon">
              <i class="pi pi-times-circle error"></i>
            </div>
            <h2>Something went wrong</h2>
            <p class="error-message">
              The AI generation failed at step: {{ session()!.currentStep }}
            </p>
            <div class="error-actions">
              <p-button
                label="Try Again"
                icon="pi pi-refresh"
                (onClick)="retryCurrentStep()"
              ></p-button>
              <p-button
                label="Start Over"
                icon="pi pi-replay"
                [outlined]="true"
                (onClick)="resetWorkflow()"
              ></p-button>
            </div>
          </div>
        </p-card>
      }

      <!-- Step 2: Topic Selection -->
      @if (session() && session()!.currentStep === 'selection' && session()!.status === 'completed') {
        <p-card header="Step 2: Select Topics">
          <p class="section-intro">
            The AI has generated topic suggestions for your course.
            Approve the ones you want to include, reject the ones you don't,
            or add your own custom topics.
          </p>

          <div class="suggestions-list">
            @for (suggestion of session()!.suggestions; track suggestion.id) {
              <div class="suggestion-card" [class.approved]="suggestion.status === 'approved'" [class.rejected]="suggestion.status === 'rejected'">
                <div class="suggestion-content">
                  <h4>{{ suggestion.title }}</h4>
                  <p>{{ suggestion.description }}</p>
                  @if (suggestion.isCustom) {
                    <span class="custom-badge">Custom</span>
                  }
                </div>
                <div class="suggestion-actions">
                  @if (suggestion.status === 'pending') {
                    <p-button
                      icon="pi pi-thumbs-up"
                      [rounded]="true"
                      severity="success"
                      [outlined]="true"
                      (onClick)="approveSuggestion(suggestion)"
                      pTooltip="Approve"
                    ></p-button>
                    <p-button
                      icon="pi pi-thumbs-down"
                      [rounded]="true"
                      severity="danger"
                      [outlined]="true"
                      (onClick)="rejectSuggestion(suggestion)"
                      pTooltip="Reject"
                    ></p-button>
                  } @else {
                    <span class="status-badge" [class]="suggestion.status">
                      {{ suggestion.status === 'approved' ? 'Approved' : 'Rejected' }}
                    </span>
                    <p-button
                      icon="pi pi-undo"
                      [rounded]="true"
                      [outlined]="true"
                      severity="secondary"
                      (onClick)="resetSuggestion(suggestion)"
                      pTooltip="Reset"
                    ></p-button>
                  }
                </div>
              </div>
            }
          </div>

          <div class="selection-actions">
            <p-button
              label="Generate More Ideas"
              icon="pi pi-plus"
              [outlined]="true"
              [loading]="generatingMore()"
              (onClick)="generateMoreSuggestions()"
            ></p-button>
            <p-button
              label="Add Custom Topic"
              icon="pi pi-pencil"
              [outlined]="true"
              (onClick)="showCustomTopicDialog = true"
            ></p-button>
          </div>

          <div class="proceed-section">
            <div class="approved-count">
              <i class="pi pi-check-circle"></i>
              {{ approvedCount() }} topics approved
            </div>
            <p-button
              label="Proceed to Refinement"
              icon="pi pi-arrow-right"
              iconPos="right"
              [disabled]="approvedCount() === 0"
              [loading]="loading()"
              (onClick)="proceedToRefinement()"
            ></p-button>
          </div>
        </p-card>
      }

      <!-- Step 3: Refinement Review (Enhanced) -->
      @if (session() && session()!.currentStep === 'script' && session()!.status === 'completed' && session()!.refinedTopics.length > 0) {
        <p-card header="Step 3: Review Refined Topics">
          <p class="section-intro">
            The AI has refined your selected topics with detailed learning goals.
            You can <strong>edit</strong> any topic, <strong>regenerate</strong> content with AI,
            or <strong>reorder</strong> the topics. The current order reflects the AI's suggested structure.
          </p>

          <!-- Reorder Mode Toggle -->
          <div class="reorder-controls">
            @if (!reorderMode()) {
              <p-button
                label="Reorder Topics"
                icon="pi pi-sort-alt"
                [outlined]="true"
                (onClick)="enterReorderMode()"
              ></p-button>
            } @else {
              <p-button
                label="Cancel"
                icon="pi pi-times"
                [outlined]="true"
                severity="secondary"
                (onClick)="cancelReorderMode()"
              ></p-button>
              <p-button
                label="Save Order"
                icon="pi pi-check"
                [loading]="savingOrder()"
                (onClick)="saveTopicOrder()"
              ></p-button>
            }
          </div>

          <!-- Reorder Mode: Move Up/Down Buttons -->
          @if (reorderMode()) {
            <div class="reorder-list">
              <p class="reorder-hint">
                <i class="pi pi-info-circle"></i>
                Use the arrows to reorder topics. Topics will be taught in this sequence.
              </p>
              @for (topic of reorderedTopics(); track topic.id; let i = $index; let first = $first; let last = $last) {
                <div class="reorder-item">
                  <div class="reorder-buttons">
                    <p-button
                      icon="pi pi-chevron-up"
                      [rounded]="true"
                      [text]="true"
                      severity="secondary"
                      [disabled]="first"
                      (onClick)="moveTopicUp(i)"
                      pTooltip="Move Up"
                    ></p-button>
                    <p-button
                      icon="pi pi-chevron-down"
                      [rounded]="true"
                      [text]="true"
                      severity="secondary"
                      [disabled]="last"
                      (onClick)="moveTopicDown(i)"
                      pTooltip="Move Down"
                    ></p-button>
                  </div>
                  <div class="reorder-item-content">
                    <span class="topic-number">{{ i + 1 }}.</span>
                    <strong>{{ topic.title }}</strong>
                    <span class="duration-badge">
                      <i class="pi pi-clock"></i> {{ topic.estimatedTimeMin }} min
                    </span>
                  </div>
                </div>
              }
            </div>
          } @else {
            <!-- Normal View: Editable Panels -->
            <div class="refined-topics-list">
              @for (topic of session()!.refinedTopics; track topic.id; let i = $index) {
                <p-panel
                  [header]="(i + 1) + '. ' + topic.title"
                  [toggleable]="true"
                  [collapsed]="editingTopicId() !== topic.id"
                  styleClass="refined-topic-panel"
                >
                  <ng-template pTemplate="icons">
                    <div class="topic-actions" (click)="$event.stopPropagation()">
                      @if (editingTopicId() !== topic.id) {
                        <p-button
                          icon="pi pi-pencil"
                          [rounded]="true"
                          [text]="true"
                          severity="secondary"
                          pTooltip="Edit Topic"
                          (onClick)="startEditingTopic(topic)"
                        ></p-button>
                        <p-button
                          icon="pi pi-refresh"
                          [rounded]="true"
                          [text]="true"
                          severity="secondary"
                          pTooltip="Regenerate with AI"
                          [loading]="isTopicRegenerating(topic.id)"
                          (onClick)="regenerateTopic(topic)"
                        ></p-button>
                      }
                    </div>
                  </ng-template>

                  @if (editingTopicId() === topic.id) {
                    <!-- Edit Mode -->
                    <form [formGroup]="editTopicForm" class="edit-topic-form">
                      <div class="form-field">
                        <label for="editTitle">Title</label>
                        <input pInputText id="editTitle" formControlName="title" class="w-full" />
                      </div>
                      <div class="form-field">
                        <label for="editDescription">Description</label>
                        <textarea
                          pInputTextarea
                          id="editDescription"
                          formControlName="description"
                          rows="3"
                          class="w-full"
                        ></textarea>
                      </div>
                      <div class="form-field">
                        <label for="editGoals">Learning Goals (press Enter to add)</label>
                        <p-chips
                          id="editGoals"
                          formControlName="learningGoals"
                          [addOnBlur]="true"
                          placeholder="Type a goal and press Enter"
                          styleClass="w-full"
                        ></p-chips>
                      </div>
                      <div class="form-field">
                        <label>Estimated Duration: {{ editTopicForm.get('estimatedTimeMin')?.value }} minutes</label>
                        <p-slider
                          formControlName="estimatedTimeMin"
                          [min]="1"
                          [max]="30"
                        ></p-slider>
                      </div>
                      <div class="edit-actions">
                        <p-button
                          label="Cancel"
                          [outlined]="true"
                          severity="secondary"
                          (onClick)="cancelEditingTopic()"
                        ></p-button>
                        <p-button
                          label="Save Changes"
                          icon="pi pi-check"
                          [disabled]="editTopicForm.invalid"
                          (onClick)="saveTopicEdit(topic)"
                        ></p-button>
                      </div>
                    </form>
                  } @else {
                    <!-- View Mode -->
                    <div class="topic-view">
                      <p class="topic-description">{{ topic.description }}</p>
                      <div class="learning-goals">
                        <strong>Learning Goals:</strong>
                        <ul>
                          @for (goal of parseLearningGoals(topic.learningGoals); track $index) {
                            <li>{{ goal }}</li>
                          }
                        </ul>
                      </div>
                      <span class="duration-badge">
                        <i class="pi pi-clock"></i> {{ topic.estimatedTimeMin }} min
                      </span>
                    </div>
                  }
                </p-panel>
              }
            </div>
          }

          <div class="proceed-section">
            @if (!reorderMode()) {
              <div class="structure-note">
                <i class="pi pi-lightbulb"></i>
                <span>The topic order above reflects the AI's recommended learning sequence</span>
              </div>
              <p-button
                label="Generate Scripts"
                icon="pi pi-file-edit"
                iconPos="right"
                [loading]="loading()"
                [disabled]="editingTopicId() !== null"
                (onClick)="proceedToScriptGeneration()"
              ></p-button>
            }
          </div>
        </p-card>
      }

      <!-- Confirm Dialog for Regeneration -->
      <p-confirmDialog></p-confirmDialog>

      <!-- Step 4: Script Review & Output Generation -->
      @if (session() && session()!.currentStep === 'video' && session()!.status === 'completed' && session()!.lessonScripts.length > 0) {
        <p-card header="Step 4: Generate Outputs">
          <p class="section-intro">
            Scripts have been generated. Choose video or presentation for each lesson,
            then generate the outputs.
          </p>

          <div class="scripts-list">
            @for (script of session()!.lessonScripts; track script.id) {
              <div class="script-card">
                <div class="script-header">
                  <h4>{{ script.title }}</h4>
                  <div class="script-meta">
                    <span class="duration-badge">
                      <i class="pi pi-clock"></i> {{ script.durationMin }} min
                    </span>
                    <p-dropdown
                      [options]="outputTypeOptions"
                      [(ngModel)]="lessonOutputTypes[script.id]"
                      (onChange)="onOutputTypeChange(script)"
                      placeholder="Output Type"
                      styleClass="output-type-dropdown">
                    </p-dropdown>
                  </div>
                </div>
                <div class="script-preview">
                  {{ getScriptPreview(script.script) }}
                </div>
                <div class="script-actions">
                  <p-button
                    label="Edit Script"
                    icon="pi pi-pencil"
                    [outlined]="true"
                    size="small"
                    (onClick)="editScript(script)"
                  ></p-button>
                  <p-button
                    icon="pi pi-refresh"
                    [rounded]="true"
                    [text]="true"
                    severity="secondary"
                    pTooltip="Regenerate Script"
                    [loading]="regeneratingScriptId() === script.id"
                    (onClick)="regenerateScript(script)"
                  ></p-button>

                  @if (lessonOutputTypes[script.id] === 'presentation') {
                    @if (script.presentationStatus === 'completed') {
                      <p-button
                        label="Preview"
                        icon="pi pi-desktop"
                        size="small"
                        (onClick)="previewPresentation(script)"
                      ></p-button>
                    } @else if (script.presentationStatus === 'processing') {
                      <p-button
                        label="Generating..."
                        icon="pi pi-spin pi-spinner"
                        size="small"
                        [disabled]="true"
                      ></p-button>
                    } @else {
                      <p-button
                        label="Generate Presentation"
                        icon="pi pi-desktop"
                        size="small"
                        [loading]="generatingPresentationId() === script.id"
                        (onClick)="generatePresentation(script)"
                      ></p-button>
                    }
                  } @else {
                    @if (script.videoStatus === 'completed') {
                      <span class="status-badge success">
                        <i class="pi pi-check"></i> Video Ready
                      </span>
                    } @else if (script.videoStatus === 'processing' || script.videoStatus === 'pending') {
                      <p-button
                        label="Generating..."
                        icon="pi pi-spin pi-spinner"
                        size="small"
                        [disabled]="true"
                      ></p-button>
                    } @else {
                      <p-button
                        label="Generate Video"
                        icon="pi pi-video"
                        size="small"
                        [loading]="generatingVideoId() === script.id"
                        (onClick)="generateVideo(script)"
                      ></p-button>
                    }
                  }
                </div>
              </div>
            }
          </div>

          <div class="proceed-section">
            <p-button
              label="Create Training"
              icon="pi pi-check"
              iconPos="right"
              [loading]="loading()"
              (onClick)="createTraining()"
              pTooltip="Finalize workflow and create the training course"
            ></p-button>
          </div>
        </p-card>
      }

      <!-- Presentation Preview Dialog -->
      <p-dialog
        header="Presentation Preview"
        [(visible)]="showPresentationDialog"
        [modal]="true"
        [style]="{ width: '90vw', height: '90vh' }"
        [maximizable]="true"
      >
        @if (currentPresentation()) {
          <app-presentation-player
            [slides]="currentPresentation()!.slides"
            [autoPlay]="true">
          </app-presentation-player>
        }
      </p-dialog>

      <!-- Step 5: Question Generation -->
      @if (session() && session()!.currentStep === 'questions' && session()!.status === 'completed') {
        <p-card header="Step 5: Generate Quiz Questions">
          <p class="section-intro">
            Videos and presentations have been generated. Now let's create quiz questions to test learners' knowledge.
          </p>

          <div class="text-center p-4">
            <p class="mb-3">Click the button below to generate quiz questions based on the course content.</p>
            <p-button
              label="Generate Questions & Complete Course"
              icon="pi pi-question-circle"
              [loading]="generatingQuestions()"
              (onClick)="proceedToQuestionGeneration()">
            </p-button>
          </div>
        </p-card>
      }

      <!-- Step 6: Completion -->
      @if (session() && session()!.currentStep === 'completed') {
        <p-card header="Course Generation Complete!">
          <div class="completion-status">
            <div class="status-icon">
              <i class="pi pi-check-circle success"></i>
            </div>
            <h2>All Done!</h2>
            <p>Your course has been generated with {{ session()!.lessonScripts.length }} lessons.</p>

            <div class="video-status-list">
              @for (script of session()!.lessonScripts; track script.id) {
                <div class="video-status-item">
                  <span>{{ script.title }}</span>
                  <span class="video-badge" [class]="script.videoStatus || 'pending'">
                    @if (script.videoStatus === 'completed') {
                      <i class="pi pi-check"></i> Ready
                    } @else if (script.videoStatus === 'failed') {
                      <i class="pi pi-times"></i> Failed
                    } @else {
                      <i class="pi pi-spin pi-spinner"></i> Generating
                    }
                  </span>
                </div>
              }
            </div>

            <div class="completion-actions">
              <a [routerLink]="['/admin']">
                <p-button label="Manage Courses" icon="pi pi-list"></p-button>
              </a>
              <p-button
                label="Create Another"
                icon="pi pi-plus"
                [outlined]="true"
                (onClick)="resetWorkflow()"
              ></p-button>
            </div>
          </div>
        </p-card>
      }

      <!-- Custom Topic Dialog -->
      <p-dialog
        header="Add Custom Topic"
        [(visible)]="showCustomTopicDialog"
        [modal]="true"
        [style]="{ width: '500px' }"
      >
        <form [formGroup]="customTopicForm" (ngSubmit)="addCustomTopic()">
          <div class="form-field">
            <label for="customTitle">Topic Title *</label>
            <input pInputText id="customTitle" formControlName="title" class="w-full" />
          </div>
          <div class="form-field">
            <label for="customDescription">Description *</label>
            <textarea
              pInputTextarea
              id="customDescription"
              formControlName="description"
              rows="3"
              class="w-full"
            ></textarea>
          </div>
          <div class="dialog-actions">
            <p-button
              label="Cancel"
              [outlined]="true"
              (onClick)="showCustomTopicDialog = false"
            ></p-button>
            <p-button
              type="submit"
              label="Add Topic"
              [disabled]="customTopicForm.invalid"
            ></p-button>
          </div>
        </form>
      </p-dialog>

      <!-- Script Editor Dialog -->
      <p-dialog
        header="Edit Lesson Script"
        [(visible)]="showScriptDialog"
        [modal]="true"
        [style]="{ width: '800px', maxHeight: '90vh' }"
      >
        @if (editingScript()) {
          <div class="script-editor">
            <div class="form-field">
              <label for="scriptTitle">Title</label>
              <input
                pInputText
                id="scriptTitle"
                [(ngModel)]="editScriptTitle"
                class="w-full"
              />
            </div>
            <div class="form-field">
              <label for="scriptContent">Script Content</label>
              <textarea
                pInputTextarea
                id="scriptContent"
                [(ngModel)]="editScriptContent"
                rows="20"
                class="w-full script-textarea"
              ></textarea>
            </div>
            <div class="dialog-actions">
              <p-button
                label="Cancel"
                [outlined]="true"
                (onClick)="cancelScriptEdit()"
              ></p-button>
              <p-button
                label="Save Changes"
                icon="pi pi-check"
                [loading]="savingScript()"
                (onClick)="saveScriptEdit()"
              ></p-button>
            </div>
          </div>
        }
      </p-dialog>
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

    .page-header h1 { margin: 0; }
    .page-header p { margin: 0.5rem 0 0; color: var(--text-secondary); }

    .workflow-steps {
      margin-bottom: 2rem;
    }

    .hint {
      color: var(--text-secondary);
      margin-top: 0.25rem;
      display: block;
    }

    .form-actions {
      margin-top: 2rem;
    }

    .section-intro {
      color: var(--text-secondary);
      margin-bottom: 1.5rem;
    }

    /* Processing Status */
    .processing-status {
      text-align: center;
      padding: 3rem;
    }

    .status-icon i {
      font-size: 4rem;
      margin-bottom: 1rem;
      color: var(--primary-color);
    }

    .status-icon .success {
      color: var(--green-500);
    }

    .status-icon .error {
      color: var(--red-500);
    }

    .processing-status h2 {
      margin: 0 0 2rem;
      color: var(--text-secondary);
    }

    /* Error Status */
    .error-status {
      text-align: center;
      padding: 3rem;
    }

    .error-status h2 {
      margin: 0 0 1rem;
    }

    .error-message {
      color: var(--text-secondary);
      margin-bottom: 2rem;
    }

    .error-actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
    }

    /* Suggestions List */
    .suggestions-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      margin-bottom: 1.5rem;
    }

    .suggestion-card {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 1.5rem;
      background: var(--surface-ground);
      border-radius: 8px;
      border: 2px solid transparent;
      transition: all 0.2s;
    }

    .suggestion-card.approved {
      border-color: var(--green-500);
      background: rgba(34, 197, 94, 0.1);
    }

    .suggestion-card.rejected {
      border-color: var(--red-500);
      background: rgba(239, 68, 68, 0.1);
      opacity: 0.6;
    }

    .suggestion-content {
      flex: 1;
    }

    .suggestion-content h4 {
      margin: 0 0 0.5rem;
    }

    .suggestion-content p {
      margin: 0;
      color: var(--text-secondary);
      font-size: 0.9rem;
    }

    .custom-badge {
      display: inline-block;
      padding: 0.125rem 0.5rem;
      background: var(--primary-color);
      color: white;
      border-radius: 4px;
      font-size: 0.75rem;
      margin-top: 0.5rem;
    }

    .suggestion-actions {
      display: flex;
      gap: 0.5rem;
      align-items: center;
    }

    .status-badge {
      padding: 0.25rem 0.75rem;
      border-radius: 4px;
      font-size: 0.875rem;
      font-weight: 500;
    }

    .status-badge.approved {
      background: rgba(34, 197, 94, 0.2);
      color: var(--green-600);
    }

    .status-badge.rejected {
      background: rgba(239, 68, 68, 0.2);
      color: var(--red-600);
    }

    .selection-actions {
      display: flex;
      gap: 1rem;
      margin-bottom: 2rem;
    }

    .proceed-section {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-top: 1.5rem;
      border-top: 1px solid var(--surface-border);
    }

    .approved-count {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--green-600);
      font-weight: 500;
    }

    /* Refined Topics */
    .refined-topics-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      margin-bottom: 1.5rem;
    }

    .refined-topic-card {
      padding: 1.5rem;
      background: var(--surface-ground);
      border-radius: 8px;
    }

    .refined-topic-card h4 {
      margin: 0 0 0.5rem;
    }

    .refined-topic-card p {
      margin: 0 0 1rem;
      color: var(--text-secondary);
    }

    .learning-goals {
      margin-bottom: 1rem;
    }

    .learning-goals ul {
      margin: 0.5rem 0 0;
      padding-left: 1.5rem;
    }

    .learning-goals li {
      margin-bottom: 0.25rem;
      color: var(--text-secondary);
    }

    .duration-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.25rem 0.75rem;
      background: var(--surface-hover);
      border-radius: 4px;
      font-size: 0.875rem;
    }

    /* Reorder Controls */
    .reorder-controls {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 1.5rem;
    }

    .reorder-hint {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--text-secondary);
      margin-bottom: 1rem;
      padding: 0.75rem 1rem;
      background: var(--surface-ground);
      border-radius: 6px;
    }

    .reorder-list {
      margin-bottom: 1.5rem;
    }

    .reorder-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0.75rem 1rem;
      background: var(--surface-ground);
      border-radius: 6px;
      margin-bottom: 0.5rem;
    }

    .reorder-buttons {
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .reorder-item-content {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      flex: 1;
    }

    .topic-number {
      color: var(--text-secondary);
      font-weight: 500;
      min-width: 1.5rem;
    }

    /* Topic Panels */
    .refined-topic-panel {
      margin-bottom: 1rem;
    }

    .topic-actions {
      display: flex;
      gap: 0.25rem;
    }

    .edit-topic-form {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .edit-actions {
      display: flex;
      gap: 0.5rem;
      justify-content: flex-end;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid var(--surface-border);
    }

    .topic-view .topic-description {
      margin: 0 0 1rem;
      color: var(--text-secondary);
    }

    .structure-note {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: var(--text-secondary);
      font-size: 0.875rem;
    }

    .structure-note i {
      color: var(--yellow-600);
    }

    /* Fix PrimeNG chips in form */
    :host ::ng-deep .p-chips {
      width: 100%;
    }

    :host ::ng-deep .p-chips-multiple-container {
      width: 100%;
    }

    /* Scripts */
    .scripts-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      margin-bottom: 1.5rem;
    }

    .script-card {
      padding: 1.5rem;
      background: var(--surface-ground);
      border-radius: 8px;
    }

    .script-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
      gap: 1rem;
    }

    .script-header h4 {
      margin: 0;
      flex: 1;
    }

    .script-meta {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .output-type-dropdown {
      min-width: 150px;
    }

    .script-actions {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      flex-wrap: wrap;
    }

    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.5rem 1rem;
      border-radius: 4px;
      font-size: 0.875rem;
    }

    .status-badge.success {
      background: var(--green-100);
      color: var(--green-700);
    }

    .status-badge.processing {
      background: var(--blue-100);
      color: var(--blue-700);
    }

    .script-preview {
      padding: 1rem;
      background: var(--surface-section);
      border-radius: 4px;
      margin-bottom: 1rem;
      font-size: 0.875rem;
      color: var(--text-secondary);
      white-space: pre-wrap;
      max-height: 100px;
      overflow: hidden;
    }

    .script-full-content {
      padding: 1rem;
      background: var(--surface-ground);
      border-radius: 8px;
      white-space: pre-wrap;
      max-height: 60vh;
      overflow-y: auto;
    }

    .script-editor .form-field {
      margin-bottom: 1.25rem;
    }

    .script-editor label {
      display: block;
      margin-bottom: 0.5rem;
      font-weight: 500;
    }

    .script-textarea {
      font-family: inherit;
      line-height: 1.6;
    }

    .w-full {
      width: 100%;
    }

    /* Completion */
    .completion-status {
      text-align: center;
      padding: 2rem;
    }

    .video-status-list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      margin: 2rem 0;
      text-align: left;
      max-width: 500px;
      margin-left: auto;
      margin-right: auto;
    }

    .video-status-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.75rem 1rem;
      background: var(--surface-ground);
      border-radius: 4px;
    }

    .video-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-size: 0.75rem;
    }

    .video-badge.completed {
      background: rgba(34, 197, 94, 0.2);
      color: var(--green-600);
    }

    .video-badge.failed {
      background: rgba(239, 68, 68, 0.2);
      color: var(--red-600);
    }

    .video-badge.pending {
      background: rgba(59, 130, 246, 0.2);
      color: var(--primary-color);
    }

    .completion-actions {
      display: flex;
      gap: 1rem;
      justify-content: center;
      margin-top: 2rem;
    }

    /* Dialog */
    .dialog-actions {
      display: flex;
      justify-content: flex-end;
      gap: 1rem;
      margin-top: 1.5rem;
    }
  `]
})
export class CourseWorkflowComponent implements OnDestroy {
  topicForm: FormGroup;
  customTopicForm: FormGroup;
  editTopicForm: FormGroup;

  session = signal<CourseWorkflowSession | null>(null);
  loading = signal(false);
  generatingMore = signal(false);
  selectedScript = signal<{ title: string; script: string } | null>(null);

  // Refinement editing state
  editingTopicId = signal<string | null>(null);
  regeneratingTopicId = signal<string | null>(null);
  reorderMode = signal(false);
  reorderedTopics = signal<RefinedTopic[]>([]);
  savingOrder = signal(false);

  showCustomTopicDialog = false;
  showScriptDialog = false;
  showPresentationDialog = false;

  // Presentation and video state
  lessonOutputTypes: { [lessonId: string]: OutputType } = {};
  generatingPresentationId = signal<string | null>(null);
  generatingVideoId = signal<string | null>(null);
  currentPresentation = signal<LessonPresentation | null>(null);
  generatingQuestions = signal(false);

  // Script editing state
  editingScript = signal<LessonScript | null>(null);
  editScriptTitle = '';
  editScriptContent = '';
  savingScript = signal(false);
  regeneratingScriptId = signal<string | null>(null);

  private pollInterval: any;
  private presentationPollInterval: any;

  outputTypeOptions = [
    { label: 'Video', value: 'video' },
    { label: 'Presentation', value: 'presentation' }
  ];

  difficultyOptions = [
    { label: 'Beginner', value: 'beginner' },
    { label: 'Intermediate', value: 'intermediate' },
    { label: 'Advanced', value: 'advanced' }
  ];

  languageOptions = [
    { label: 'English', value: 'en' },
    { label: 'Deutsch', value: 'de' },
    { label: 'Français', value: 'fr' },
    { label: 'Español', value: 'es' },
    { label: 'Italiano', value: 'it' },
    { label: 'Português', value: 'pt' }
  ];

  workflowSteps: MenuItem[] = [
    { label: 'Topic' },
    { label: 'Selection' },
    { label: 'Refinement' },
    { label: 'Scripts' },
    { label: 'Videos' },
    { label: 'Questions' }
  ];

  currentStepIndex = computed(() => {
    const step = this.session()?.currentStep;
    if (!step) return 0;
    const stepMap: Record<WorkflowStep, number> = {
      'research': 0,
      'selection': 1,
      'refinement': 2,
      'script': 3,
      'video': 4,
      'questions': 5,
      'completed': 5
    };
    return stepMap[step] ?? 0;
  });

  approvedCount = computed(() => {
    return this.session()?.suggestions.filter(s => s.status === 'approved').length ?? 0;
  });

  constructor(
    private fb: FormBuilder,
    private workflowService: WorkflowService,
    private messageService: MessageService,
    private confirmationService: ConfirmationService,
    private router: Router
  ) {
    this.topicForm = this.fb.group({
      topic: ['', [Validators.required, Validators.minLength(5), Validators.maxLength(500)]],
      targetAudience: [''],
      difficultyLevel: ['intermediate'],
      language: ['en', Validators.required],
      videoDurationMin: [5]
    });

    this.customTopicForm = this.fb.group({
      title: ['', [Validators.required, Validators.minLength(3)]],
      description: ['', [Validators.required, Validators.minLength(10)]]
    });

    this.editTopicForm = this.fb.group({
      title: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(200)]],
      description: ['', [Validators.required, Validators.maxLength(2000)]],
      learningGoals: [[], [Validators.required]],
      estimatedTimeMin: [10, [Validators.required, Validators.min(1), Validators.max(60)]]
    });
  }

  ngOnDestroy(): void {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
    if (this.presentationPollInterval) {
      clearInterval(this.presentationPollInterval);
    }
  }

  isProcessing(): boolean {
    const session = this.session();
    if (!session) return false;
    return session.status === 'processing' || session.status === 'pending';
  }

  getProcessingTitle(): string {
    const step = this.session()?.currentStep;
    switch (step) {
      case 'research': return 'Step 1: Researching Topics';
      case 'refinement': return 'Step 3: Refining Topics';
      case 'script': return 'Step 4: Generating Scripts';
      case 'video': return 'Step 5: Generating Videos';
      case 'questions': return 'Step 6: Generating Questions';
      default: return 'Processing...';
    }
  }

  getProcessingMessage(): string {
    const step = this.session()?.currentStep;
    switch (step) {
      case 'research': return 'AI is researching and generating topic suggestions...';
      case 'refinement': return 'AI is refining your selected topics with learning goals...';
      case 'script': return 'AI is writing detailed scripts for each lesson...';
      case 'video': return 'Videos are being generated with Synthesia...';
      default: return 'Please wait...';
    }
  }

  startResearch(): void {
    if (this.topicForm.invalid) return;

    this.loading.set(true);
    this.workflowService.startResearch(this.topicForm.value).subscribe({
      next: (session) => {
        this.session.set(session);
        this.loading.set(false);
        this.startPolling();
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not start research'
        });
      }
    });
  }

  startPolling(): void {
    // Clear any existing polling interval first
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }

    this.pollInterval = setInterval(() => {
      const sessionId = this.session()?.id;
      if (!sessionId) {
        clearInterval(this.pollInterval);
        return;
      }

      this.workflowService.getSession(sessionId).subscribe({
        next: (session) => {
          if (session && session.id) {
            this.session.set(session);
            this.initializeLessonOutputTypes();
            // Stop polling when we reach a stable state (completed status at certain steps)
            if (session.status === 'completed' || session.status === 'failed') {
              // Stop at interactive steps or completion
              if (session.currentStep === 'completed' || session.currentStep === 'selection' ||
                  session.currentStep === 'script' || session.currentStep === 'video' ||
                  session.currentStep === 'questions') {
                clearInterval(this.pollInterval);
                this.pollInterval = null;
              }
            }
          }
        },
        error: (err) => {
          console.error('Polling error:', err);
          // Don't clear session on polling error, just stop polling
          clearInterval(this.pollInterval);
          this.pollInterval = null;
        }
      });
    }, 2000);
  }

  approveSuggestion(suggestion: TopicSuggestion): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.updateSuggestionStatus(sessionId, suggestion.id, 'approved').subscribe({
      next: () => {
        this.refreshSession();
      }
    });
  }

  rejectSuggestion(suggestion: TopicSuggestion): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.updateSuggestionStatus(sessionId, suggestion.id, 'rejected').subscribe({
      next: () => {
        this.refreshSession();
      }
    });
  }

  resetSuggestion(suggestion: TopicSuggestion): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.updateSuggestionStatus(sessionId, suggestion.id, 'pending').subscribe({
      next: () => {
        this.refreshSession();
      }
    });
  }

  generateMoreSuggestions(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.generatingMore.set(true);
    this.workflowService.generateMoreSuggestions(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.generatingMore.set(false);
      },
      error: (err) => {
        this.generatingMore.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not generate more suggestions'
        });
      }
    });
  }

  addCustomTopic(): void {
    if (this.customTopicForm.invalid) return;

    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.addCustomTopic(sessionId, this.customTopicForm.value).subscribe({
      next: () => {
        this.showCustomTopicDialog = false;
        this.customTopicForm.reset();
        this.refreshSession();
        this.messageService.add({
          severity: 'success',
          summary: 'Topic Added',
          detail: 'Your custom topic has been added and approved'
        });
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not add custom topic'
        });
      }
    });
  }

  proceedToRefinement(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    // Clear any existing polling
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }

    this.loading.set(true);
    this.workflowService.proceedToRefinement(sessionId).subscribe({
      next: (session) => {
        console.log('Refinement started, session:', session);
        if (session && session.id) {
          this.session.set(session);
          this.startPolling();
        } else {
          console.error('Invalid session response:', session);
          this.messageService.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Invalid response from server'
          });
        }
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Refinement error:', err);
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not proceed to refinement'
        });
      }
    });
  }

  proceedToScriptGeneration(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.loading.set(true);
    this.workflowService.proceedToScriptGeneration(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.loading.set(false);
        this.startPolling();
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not proceed to script generation'
        });
      }
    });
  }

  proceedToVideoGeneration(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.loading.set(true);
    this.workflowService.proceedToVideoGeneration(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.loading.set(false);
        this.startPolling();
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not proceed to video generation'
        });
      }
    });
  }

  proceedToQuestionGeneration(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.generatingQuestions.set(true);
    this.workflowService.proceedToQuestionGeneration(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.generatingQuestions.set(false);
        this.startPolling();
        this.messageService.add({
          severity: 'info',
          summary: 'Generating',
          detail: 'Quiz questions are being generated...'
        });
      },
      error: (err) => {
        this.generatingQuestions.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not generate questions'
        });
      }
    });
  }

  refreshSession(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.getSession(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.initializeLessonOutputTypes();
      }
    });
  }

  parseLearningGoals(goals: any): string[] {
    if (Array.isArray(goals)) return goals;
    if (typeof goals === 'string') {
      try {
        return JSON.parse(goals);
      } catch {
        return [goals];
      }
    }
    return [];
  }

  getScriptPreview(script: string): string {
    return script.length > 200 ? script.substring(0, 200) + '...' : script;
  }

  viewScript(script: { title: string; script: string }): void {
    this.selectedScript.set(script);
    this.showScriptDialog = true;
  }

  editScript(script: LessonScript): void {
    this.editingScript.set(script);
    this.editScriptTitle = script.title;
    this.editScriptContent = script.script;
    this.showScriptDialog = true;
  }

  cancelScriptEdit(): void {
    this.editingScript.set(null);
    this.editScriptTitle = '';
    this.editScriptContent = '';
    this.showScriptDialog = false;
  }

  saveScriptEdit(): void {
    const script = this.editingScript();
    const sessionId = this.session()?.id;
    if (!script || !sessionId) return;

    this.savingScript.set(true);

    const request: UpdateLessonScriptRequest = {
      title: this.editScriptTitle,
      script: this.editScriptContent
    };

    this.workflowService.updateLessonScript(sessionId, script.id, request).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Saved',
          detail: 'Script has been updated'
        });
        this.savingScript.set(false);
        this.showScriptDialog = false;
        this.editingScript.set(null);
        this.refreshSession();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not save script'
        });
        this.savingScript.set(false);
      }
    });
  }

  regenerateScript(script: LessonScript): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to regenerate the script for "${script.title}"? This will replace the current script with new AI-generated content.`,
      header: 'Confirm Regeneration',
      icon: 'pi pi-refresh',
      accept: () => {
        this.performScriptRegeneration(script);
      }
    });
  }

  private performScriptRegeneration(script: LessonScript): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.regeneratingScriptId.set(script.id);

    this.workflowService.regenerateScript(sessionId, script.id).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Regenerated',
          detail: 'Script has been regenerated'
        });
        this.regeneratingScriptId.set(null);
        this.refreshSession();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not regenerate script'
        });
        this.regeneratingScriptId.set(null);
      }
    });
  }

  retryCurrentStep(): void {
    const session = this.session();
    if (!session) return;

    // Retry based on the current step that failed
    switch (session.currentStep) {
      case 'refinement':
        this.proceedToRefinement();
        break;
      case 'script':
        this.proceedToScriptGeneration();
        break;
      case 'video':
        this.proceedToVideoGeneration();
        break;
      case 'questions':
        this.proceedToQuestionGeneration();
        break;
      default:
        // For other steps, just refresh the session
        this.refreshSession();
    }
  }

  resetWorkflow(): void {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
    this.session.set(null);
    this.topicForm.reset({
      difficultyLevel: 'intermediate',
      videoDurationMin: 5
    });
  }

  // ===== REFINEMENT EDITING METHODS =====

  startEditingTopic(topic: RefinedTopic): void {
    this.editingTopicId.set(topic.id);
    const goals = this.parseLearningGoals(topic.learningGoals);
    this.editTopicForm.patchValue({
      title: topic.title,
      description: topic.description,
      learningGoals: goals,
      estimatedTimeMin: topic.estimatedTimeMin
    });
  }

  cancelEditingTopic(): void {
    this.editingTopicId.set(null);
    this.editTopicForm.reset();
  }

  saveTopicEdit(topic: RefinedTopic): void {
    if (this.editTopicForm.invalid) return;

    const sessionId = this.session()?.id;
    if (!sessionId) return;

    const request: UpdateRefinedTopicRequest = {
      title: this.editTopicForm.value.title,
      description: this.editTopicForm.value.description,
      learningGoals: this.editTopicForm.value.learningGoals,
      estimatedTimeMin: this.editTopicForm.value.estimatedTimeMin
    };

    this.workflowService.updateRefinedTopic(sessionId, topic.id, request).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Topic Updated',
          detail: 'Your changes have been saved'
        });
        this.editingTopicId.set(null);
        this.refreshSession();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not update topic'
        });
      }
    });
  }

  regenerateTopic(topic: RefinedTopic): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to regenerate "${topic.title}"? This will replace the current content with new AI-generated content.`,
      header: 'Confirm Regeneration',
      icon: 'pi pi-refresh',
      accept: () => {
        this.performTopicRegeneration(topic);
      }
    });
  }

  private performTopicRegeneration(topic: RefinedTopic): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.regeneratingTopicId.set(topic.id);

    this.workflowService.regenerateTopic(sessionId, topic.id).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Topic Regenerated',
          detail: 'New content has been generated'
        });
        this.regeneratingTopicId.set(null);
        this.refreshSession();
      },
      error: (err) => {
        this.regeneratingTopicId.set(null);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not regenerate topic'
        });
      }
    });
  }

  isTopicRegenerating(topicId: string): boolean {
    return this.regeneratingTopicId() === topicId;
  }

  // ===== REORDERING METHODS =====

  enterReorderMode(): void {
    const topics = this.session()?.refinedTopics || [];
    this.reorderedTopics.set([...topics]);
    this.reorderMode.set(true);
  }

  cancelReorderMode(): void {
    this.reorderMode.set(false);
    this.reorderedTopics.set([]);
  }

  saveTopicOrder(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    const topics = this.reorderedTopics();
    const request: ReorderTopicsRequest = {
      topicOrders: topics.map((t, index) => ({
        topicId: t.id,
        sortOrder: index
      }))
    };

    this.savingOrder.set(true);

    this.workflowService.reorderTopics(sessionId, request).subscribe({
      next: (session) => {
        this.session.set(session);
        this.messageService.add({
          severity: 'success',
          summary: 'Order Saved',
          detail: 'Topic order has been updated'
        });
        this.savingOrder.set(false);
        this.reorderMode.set(false);
        this.reorderedTopics.set([]);
      },
      error: (err) => {
        this.savingOrder.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not save order'
        });
      }
    });
  }

  moveTopicUp(index: number): void {
    if (index <= 0) return;
    const topics = [...this.reorderedTopics()];
    [topics[index - 1], topics[index]] = [topics[index], topics[index - 1]];
    this.reorderedTopics.set(topics);
  }

  moveTopicDown(index: number): void {
    const topics = [...this.reorderedTopics()];
    if (index >= topics.length - 1) return;
    [topics[index], topics[index + 1]] = [topics[index + 1], topics[index]];
    this.reorderedTopics.set(topics);
  }

  // Initialize lesson output types when session changes
  initializeLessonOutputTypes(): void {
    const session = this.session();
    if (session?.lessonScripts) {
      for (const script of session.lessonScripts) {
        if (!this.lessonOutputTypes[script.id]) {
          this.lessonOutputTypes[script.id] = script.outputType || 'video';
        }
      }
    }
  }

  onOutputTypeChange(script: LessonScript): void {
    const outputType = this.lessonOutputTypes[script.id];
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.setOutputType(sessionId, script.id, outputType).subscribe({
      next: (session) => {
        this.session.set(session);
        this.messageService.add({
          severity: 'success',
          summary: 'Updated',
          detail: `Output type set to ${outputType}`
        });
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not update output type'
        });
      }
    });
  }

  generatePresentation(script: LessonScript): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.generatingPresentationId.set(script.id);

    this.workflowService.generatePresentation(sessionId, script.id).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'info',
          summary: 'Generating',
          detail: 'Presentation is being generated. This may take a few moments...'
        });
        // Start polling for presentation completion
        this.startPresentationPolling(script.id);
      },
      error: (err) => {
        this.generatingPresentationId.set(null);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not generate presentation'
        });
      }
    });
  }

  private startPresentationPolling(lessonId: string): void {
    // Clear any existing presentation polling
    if (this.presentationPollInterval) {
      clearInterval(this.presentationPollInterval);
      this.presentationPollInterval = null;
    }

    this.presentationPollInterval = setInterval(() => {
      const sessionId = this.session()?.id;
      if (!sessionId) {
        clearInterval(this.presentationPollInterval);
        this.presentationPollInterval = null;
        return;
      }

      this.workflowService.getSession(sessionId).subscribe({
        next: (session) => {
          this.session.set(session);
          this.initializeLessonOutputTypes();

          // Find the lesson we're generating for
          const lesson = session.lessonScripts?.find(s => s.id === lessonId);
          if (lesson) {
            const status = lesson.presentationStatus;
            if (status === 'completed') {
              clearInterval(this.presentationPollInterval);
              this.presentationPollInterval = null;
              this.generatingPresentationId.set(null);
              this.messageService.add({
                severity: 'success',
                summary: 'Completed',
                detail: 'Presentation has been generated successfully!'
              });
            } else if (status === 'failed') {
              clearInterval(this.presentationPollInterval);
              this.presentationPollInterval = null;
              this.generatingPresentationId.set(null);
              this.messageService.add({
                severity: 'error',
                summary: 'Failed',
                detail: 'Presentation generation failed. Please try again.'
              });
            }
            // If still 'processing', continue polling
          }
        },
        error: () => {
          // Don't stop polling on temporary errors
          console.error('Presentation polling error');
        }
      });
    }, 3000); // Poll every 3 seconds
  }

  previewPresentation(script: LessonScript): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.workflowService.getPresentation(sessionId, script.id).subscribe({
      next: (presentation) => {
        this.currentPresentation.set(presentation);
        this.showPresentationDialog = true;
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not load presentation'
        });
      }
    });
  }

  generateVideo(script: LessonScript): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.generatingVideoId.set(script.id);

    // For now, we'll use the proceedToVideoGeneration which handles all videos
    // In the future, we could add an API endpoint to generate a single video
    this.workflowService.proceedToVideoGeneration(sessionId).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'info',
          summary: 'Generating',
          detail: 'Video generation has started...'
        });
        this.startVideoPolling(script.id);
      },
      error: (err) => {
        this.generatingVideoId.set(null);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not generate video'
        });
      }
    });
  }

  private startVideoPolling(lessonId: string): void {
    // Use the existing polling mechanism
    this.startPolling();

    // Also check specifically for video completion
    const checkInterval = setInterval(() => {
      const session = this.session();
      if (!session) {
        clearInterval(checkInterval);
        return;
      }

      const lesson = session.lessonScripts?.find(s => s.id === lessonId);
      if (lesson) {
        if (lesson.videoStatus === 'completed') {
          clearInterval(checkInterval);
          this.generatingVideoId.set(null);
          this.messageService.add({
            severity: 'success',
            summary: 'Completed',
            detail: 'Video has been generated successfully!'
          });
        } else if (lesson.videoStatus === 'failed') {
          clearInterval(checkInterval);
          this.generatingVideoId.set(null);
          this.messageService.add({
            severity: 'error',
            summary: 'Failed',
            detail: 'Video generation failed. Please try again.'
          });
        }
      }
    }, 5000);
  }

  createTraining(): void {
    const sessionId = this.session()?.id;
    if (!sessionId) return;

    this.loading.set(true);
    this.workflowService.proceedToVideoGeneration(sessionId).subscribe({
      next: (session) => {
        this.session.set(session);
        this.loading.set(false);
        this.startPolling();
        this.messageService.add({
          severity: 'success',
          summary: 'Training Created',
          detail: 'Your training course is being finalized. Videos are being generated for video-type lessons.'
        });
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: err.error?.error || 'Could not create training'
        });
      }
    });
  }

}
