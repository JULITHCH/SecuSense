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
  createdAt: string;
}

export interface Test {
  id: string;
  courseId: string;
  title: string;
  description: string;
  timeLimitMinutes?: number;
  passingScore: number;
  createdAt: string;
  updatedAt: string;
  questions?: Question[];
}

export interface CreateQuestionRequest {
  testId: string;
  questionType: QuestionType;
  questionText: string;
  questionData: any;
  points: number;
  orderIndex?: number;
}

export interface UpdateQuestionRequest {
  questionType: QuestionType;
  questionText: string;
  questionData: any;
  points: number;
}

@Injectable({
  providedIn: 'root'
})
export class QuestionService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getTestWithQuestions(courseId: string): Observable<Test> {
    return this.http.get<Test>(`${this.API_URL}/admin/courses/${courseId}/questions`);
  }

  createQuestion(request: CreateQuestionRequest): Observable<Question> {
    return this.http.post<Question>(`${this.API_URL}/admin/questions`, request);
  }

  updateQuestion(questionId: string, request: UpdateQuestionRequest): Observable<Question> {
    return this.http.put<Question>(`${this.API_URL}/admin/questions/${questionId}`, request);
  }

  deleteQuestion(questionId: string): Observable<any> {
    return this.http.delete(`${this.API_URL}/admin/questions/${questionId}`);
  }
}
