import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export interface GenerateCourseRequest {
  topic: string;
  targetAudience?: string;
  difficultyLevel?: 'beginner' | 'intermediate' | 'advanced';
  videoDurationMin?: number;
}

export interface AIGenerationJob {
  id: string;
  courseId?: string;
  jobType: 'content_generation' | 'video_generation' | 'test_generation';
  status: 'pending' | 'processing' | 'completed' | 'failed';
  inputData: any;
  outputData?: any;
  error?: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

@Injectable({
  providedIn: 'root'
})
export class AIService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  generateCourse(request: GenerateCourseRequest): Observable<AIGenerationJob> {
    return this.http.post<AIGenerationJob>(`${this.API_URL}/admin/generate/course`, request);
  }

  getJob(jobId: string): Observable<AIGenerationJob> {
    return this.http.get<AIGenerationJob>(`${this.API_URL}/admin/generate/jobs/${jobId}`);
  }
}
