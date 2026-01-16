import { Routes } from '@angular/router';

export const COURSES_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./pages/catalog/catalog.component').then(m => m.CatalogComponent)
  },
  {
    path: ':id',
    loadComponent: () => import('./pages/detail/detail.component').then(m => m.CourseDetailComponent)
  }
];
