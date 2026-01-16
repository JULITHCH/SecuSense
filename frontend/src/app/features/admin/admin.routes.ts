import { Routes } from '@angular/router';

export const ADMIN_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./pages/course-list/course-list.component').then(m => m.AdminCourseListComponent)
  },
  {
    path: 'generate',
    loadComponent: () => import('./pages/course-generator/course-generator.component').then(m => m.CourseGeneratorComponent)
  }
];
