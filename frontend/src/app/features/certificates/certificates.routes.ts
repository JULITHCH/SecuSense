import { Routes } from '@angular/router';

export const CERTIFICATES_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./pages/list/list.component').then(m => m.CertificateListComponent)
  },
  {
    path: 'verify/:hash',
    loadComponent: () => import('./pages/verify/verify.component').then(m => m.VerifyComponent)
  },
  {
    path: ':id',
    loadComponent: () => import('./pages/list/list.component').then(m => m.CertificateListComponent)
  }
];
