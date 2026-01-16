import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';

export interface Certificate {
  id: string;
  certificateNumber: string;
  userId: string;
  courseId: string;
  testAttemptId: string;
  pdfUrl?: string;
  verificationHash: string;
  issuedAt: string;
  expiresAt?: string;
  userFirstName?: string;
  userLastName?: string;
  userEmail?: string;
  courseTitle?: string;
  score?: number;
  maxScore?: number;
}

export interface CertificateVerification {
  valid: boolean;
  certificateNumber?: string;
  holderName?: string;
  courseTitle?: string;
  issuedAt?: string;
  score?: number;
  maxScore?: number;
}

export interface GenerateCertificateRequest {
  courseId: string;
  attemptId: string;
}

@Injectable({
  providedIn: 'root'
})
export class CertificateService {
  private readonly API_URL = environment.apiUrl;

  constructor(private http: HttpClient) {}

  getMyCertificates(): Observable<Certificate[]> {
    return this.http.get<Certificate[]>(`${this.API_URL}/certificates`);
  }

  getCertificateById(id: string): Observable<Certificate> {
    return this.http.get<Certificate>(`${this.API_URL}/certificates/${id}`);
  }

  generateCertificate(request: GenerateCertificateRequest): Observable<Certificate> {
    return this.http.post<Certificate>(`${this.API_URL}/certificates`, request);
  }

  downloadCertificate(id: string): Observable<Blob> {
    return this.http.get(`${this.API_URL}/certificates/${id}/download`, {
      responseType: 'blob'
    });
  }

  verifyCertificate(hash: string): Observable<CertificateVerification> {
    return this.http.get<CertificateVerification>(`${this.API_URL}/certificates/verify/${hash}`);
  }
}
