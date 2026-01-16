import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export type QuestionType = 'multiple_choice' | 'drag_drop' | 'fill_blank' | 'matching' | 'ordering';

export interface Question {
  id: string;
  testId: string;
  questionType: QuestionType;
  questionText: string;
  questionData: any;
  points: number;
  orderIndex: number;
}

export interface Test {
  id: string;
  courseId: string;
  title: string;
  description: string;
  timeLimitMinutes?: number;
  passingScore: number;
  questions?: Question[];
}

export interface TestAttempt {
  id: string;
  userId: string;
  testId: string;
  startedAt: string;
  completedAt?: string;
  score?: number;
  maxScore?: number;
  percentage?: number;
  passed?: boolean;
}

export interface SubmitAnswer {
  questionId: string;
  answerData: any;
}

export interface AnswerResult {
  questionId: string;
  isCorrect: boolean;
  pointsAwarded: number;
  maxPoints: number;
  explanation?: string;
}

export interface TestResult {
  attemptId: string;
  score: number;
  maxScore: number;
  percentage: number;
  passed: boolean;
  answers: AnswerResult[];
}

@Injectable({
  providedIn: 'root'
})
export class TestService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getTestByCourseId(courseId: string): Observable<Test> {
    return this.http.get<Test>(`${this.API_URL}/courses/${courseId}/test`);
  }

  startAttempt(testId: string): Observable<TestAttempt> {
    return this.http.post<TestAttempt>(`${this.API_URL}/tests/${testId}/attempts`, {});
  }

  submitAttempt(attemptId: string, answers: SubmitAnswer[]): Observable<TestResult> {
    return this.http.post<TestResult>(`${this.API_URL}/attempts/${attemptId}/submit`, { answers });
  }

  getAttemptResults(attemptId: string): Observable<TestResult> {
    return this.http.get<TestResult>(`${this.API_URL}/attempts/${attemptId}/results`);
  }
}
