import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export type VideoStatus = 'pending' | 'processing' | 'completed' | 'failed';

export interface Course {
  id: string;
  title: string;
  description: string;
  videoUrl?: string;
  synthesiaVideoId?: string;
  videoStatus?: VideoStatus;
  videoError?: string;
  thumbnailUrl?: string;
  passPercentage: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface CreateCourseRequest {
  title: string;
  description: string;
  passPercentage: number;
}

export interface UpdateCourseRequest {
  title?: string;
  description?: string;
  videoUrl?: string;
  thumbnailUrl?: string;
  passPercentage?: number;
  isPublished?: boolean;
}

export interface PresentationSlide {
  title: string;
  content: string;
  script: string;
  audioUrl: string;
}

export interface LessonPresentation {
  id: string;
  lessonId: string;
  slides: PresentationSlide[];
  status: string;
  createdAt: string;
}

export type OutputType = 'video' | 'presentation';

export interface CourseLesson {
  id: string;
  title: string;
  outputType: OutputType;
  videoUrl?: string;
  videoStatus?: string;
  presentationStatus?: string;
  presentation?: LessonPresentation;
}

@Injectable({
  providedIn: 'root'
})
export class CourseService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getCourses(page = 1, pageSize = 20): Observable<PaginatedResponse<Course>> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('pageSize', pageSize.toString());

    return this.http.get<PaginatedResponse<Course>>(`${this.API_URL}/courses`, { params });
  }

  getCourseById(id: string): Observable<Course> {
    return this.http.get<Course>(`${this.API_URL}/courses/${id}`);
  }

  getCourseLessons(courseId: string): Observable<CourseLesson[]> {
    return this.http.get<CourseLesson[]>(`${this.API_URL}/courses/${courseId}/lessons`);
  }

  // Admin methods
  getAllCourses(page = 1, pageSize = 20): Observable<PaginatedResponse<Course>> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('pageSize', pageSize.toString());

    return this.http.get<PaginatedResponse<Course>>(`${this.API_URL}/admin/courses`, { params });
  }

  createCourse(course: CreateCourseRequest): Observable<Course> {
    return this.http.post<Course>(`${this.API_URL}/admin/courses`, course);
  }

  updateCourse(id: string, course: UpdateCourseRequest): Observable<Course> {
    return this.http.put<Course>(`${this.API_URL}/admin/courses/${id}`, course);
  }

  deleteCourse(id: string): Observable<void> {
    return this.http.delete<void>(`${this.API_URL}/admin/courses/${id}`);
  }

  publishCourse(id: string): Observable<{ published: boolean }> {
    return this.http.post<{ published: boolean }>(`${this.API_URL}/admin/courses/${id}/publish`, {});
  }

  unpublishCourse(id: string): Observable<{ published: boolean }> {
    return this.http.post<{ published: boolean }>(`${this.API_URL}/admin/courses/${id}/unpublish`, {});
  }

  refreshVideoStatus(id: string): Observable<Course> {
    return this.http.post<Course>(`${this.API_URL}/admin/courses/${id}/refresh-video`, {});
  }
}
