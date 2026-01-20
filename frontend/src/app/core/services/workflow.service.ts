import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export type WorkflowStep = 'research' | 'selection' | 'refinement' | 'script' | 'video' | 'questions' | 'completed';
export type JobStatus = 'pending' | 'processing' | 'completed' | 'failed';
export type SuggestionStatus = 'pending' | 'approved' | 'rejected';
export type OutputType = 'video' | 'presentation';

export type CourseLanguage = 'en' | 'de' | 'fr' | 'es' | 'it' | 'pt';

export interface StartResearchRequest {
  topic: string;
  targetAudience?: string;
  difficultyLevel?: 'beginner' | 'intermediate' | 'advanced';
  language: CourseLanguage;
  videoDurationMin?: number;
}

export interface TopicSuggestion {
  id: string;
  sessionId: string;
  title: string;
  description: string;
  isCustom: boolean;
  status: SuggestionStatus;
  sortOrder: number;
  createdAt: string;
}

export interface RefinedTopic {
  id: string;
  sessionId: string;
  suggestionId: string;
  title: string;
  description: string;
  learningGoals: string[];
  estimatedTimeMin: number;
  sortOrder: number;
  createdAt: string;
}

export interface LessonScript {
  id: string;
  sessionId: string;
  topicId: string;
  title: string;
  script: string;
  durationMin: number;
  outputType: OutputType;
  videoId?: string;
  videoUrl?: string;
  videoStatus?: string;
  presentationStatus?: string;
  sortOrder: number;
  createdAt: string;
}

export interface PresentationSlide {
  title: string;
  content: string;
  script: string;
  audioUrl: string;
  imageUrl?: string;
  imageAlt?: string;
  imageKeywords?: string;
}

export interface LessonPresentation {
  id: string;
  lessonId: string;
  slides: PresentationSlide[];
  status: string;
  createdAt: string;
}

export interface SetOutputTypeRequest {
  outputType: OutputType;
}

export interface CourseWorkflowSession {
  id: string;
  mainTopic: string;
  targetAudience: string;
  difficultyLevel: string;
  language: CourseLanguage;
  videoDurationMin: number;
  currentStep: WorkflowStep;
  status: JobStatus;
  courseId?: string;
  suggestions: TopicSuggestion[];
  refinedTopics: RefinedTopic[];
  lessonScripts: LessonScript[];
  createdAt: string;
  updatedAt: string;
}

export interface AddCustomTopicRequest {
  title: string;
  description: string;
}

export interface UpdateSuggestionRequest {
  status: SuggestionStatus;
}

export interface UpdateRefinedTopicRequest {
  title: string;
  description: string;
  learningGoals: string[];
  estimatedTimeMin: number;
}

export interface TopicOrder {
  topicId: string;
  sortOrder: number;
}

export interface ReorderTopicsRequest {
  topicOrders: TopicOrder[];
}

export interface UpdateLessonScriptRequest {
  title?: string;
  script: string;
}

export interface GeneratedQuestion {
  questionType: 'multiple_choice' | 'drag_drop' | 'fill_blank' | 'matching' | 'ordering';
  questionText: string;
  questionData: any;
  points: number;
}

export interface QuestionsPreview {
  questions: GeneratedQuestion[];
}

@Injectable({
  providedIn: 'root'
})
export class WorkflowService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  startResearch(request: StartResearchRequest): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/start`, request);
  }

  getSession(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.get<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/${sessionId}`);
  }

  updateSuggestionStatus(sessionId: string, suggestionId: string, status: SuggestionStatus): Observable<any> {
    return this.http.put(`${this.API_URL}/admin/workflow/${sessionId}/suggestions/${suggestionId}`, { status });
  }

  addCustomTopic(sessionId: string, request: AddCustomTopicRequest): Observable<TopicSuggestion> {
    return this.http.post<TopicSuggestion>(`${this.API_URL}/admin/workflow/${sessionId}/suggestions`, request);
  }

  generateMoreSuggestions(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/${sessionId}/generate-more`, {});
  }

  proceedToRefinement(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/${sessionId}/refine`, {});
  }

  proceedToScriptGeneration(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/${sessionId}/scripts`, {});
  }

  proceedToVideoGeneration(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(`${this.API_URL}/admin/workflow/${sessionId}/videos`, {});
  }

  updateRefinedTopic(sessionId: string, topicId: string, request: UpdateRefinedTopicRequest): Observable<RefinedTopic> {
    return this.http.put<RefinedTopic>(
      `${this.API_URL}/admin/workflow/${sessionId}/topics/${topicId}`,
      request
    );
  }

  regenerateTopic(sessionId: string, topicId: string): Observable<RefinedTopic> {
    return this.http.post<RefinedTopic>(
      `${this.API_URL}/admin/workflow/${sessionId}/topics/${topicId}/regenerate`,
      {}
    );
  }

  reorderTopics(sessionId: string, request: ReorderTopicsRequest): Observable<CourseWorkflowSession> {
    return this.http.put<CourseWorkflowSession>(
      `${this.API_URL}/admin/workflow/${sessionId}/topics/reorder`,
      request
    );
  }

  setOutputType(sessionId: string, lessonId: string, outputType: OutputType): Observable<CourseWorkflowSession> {
    return this.http.put<CourseWorkflowSession>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}/output-type`,
      { outputType }
    );
  }

  generatePresentation(sessionId: string, lessonId: string): Observable<LessonPresentation> {
    return this.http.post<LessonPresentation>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}/presentation`,
      {}
    );
  }

  getPresentation(sessionId: string, lessonId: string): Observable<LessonPresentation> {
    return this.http.get<LessonPresentation>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}/presentation`
    );
  }

  regenerateAudio(sessionId: string, lessonId: string): Observable<LessonPresentation> {
    return this.http.post<LessonPresentation>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}/regenerate-audio`,
      {}
    );
  }

  updateLessonScript(sessionId: string, lessonId: string, request: UpdateLessonScriptRequest): Observable<LessonScript> {
    return this.http.put<LessonScript>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}`,
      request
    );
  }

  regenerateScript(sessionId: string, lessonId: string): Observable<LessonScript> {
    return this.http.post<LessonScript>(
      `${this.API_URL}/admin/workflow/${sessionId}/lessons/${lessonId}/regenerate`,
      {}
    );
  }

  proceedToQuestionGeneration(sessionId: string): Observable<CourseWorkflowSession> {
    return this.http.post<CourseWorkflowSession>(
      `${this.API_URL}/admin/workflow/${sessionId}/questions`,
      {}
    );
  }

  previewQuestions(sessionId: string): Observable<QuestionsPreview> {
    return this.http.get<QuestionsPreview>(
      `${this.API_URL}/admin/workflow/${sessionId}/questions/preview`
    );
  }
}
