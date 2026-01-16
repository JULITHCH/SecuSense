import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export interface Enrollment {
  id: string;
  userId: string;
  courseId: string;
  status: 'active' | 'completed' | 'dropped';
  progressPercentage: number;
  videoWatched: boolean;
  enrolledAt: string;
  completedAt?: string;
  updatedAt: string;
}

export interface EnrollmentWithCourse extends Enrollment {
  courseTitle: string;
  courseDescription: string;
  courseThumbnailUrl?: string;
}

@Injectable({
  providedIn: 'root'
})
export class EnrollmentService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  enrollInCourse(courseId: string): Observable<Enrollment> {
    return this.http.post<Enrollment>(`${this.API_URL}/courses/${courseId}/enroll`, {});
  }

  getMyEnrollments(): Observable<EnrollmentWithCourse[]> {
    return this.http.get<EnrollmentWithCourse[]>(`${this.API_URL}/enrollments`);
  }

  getEnrollmentById(id: string): Observable<Enrollment> {
    return this.http.get<Enrollment>(`${this.API_URL}/enrollments/${id}`);
  }

  updateProgress(enrollmentId: string, progressPercentage: number): Observable<Enrollment> {
    return this.http.put<Enrollment>(`${this.API_URL}/enrollments/${enrollmentId}/progress`, {
      progressPercentage
    });
  }

  completeVideo(enrollmentId: string): Observable<Enrollment> {
    return this.http.post<Enrollment>(`${this.API_URL}/enrollments/${enrollmentId}/complete-video`, {});
  }
}
