import { Routes } from '@angular/router';

export const ADMIN_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./pages/course-list/course-list.component').then(m => m.AdminCourseListComponent)
  },
  {
    path: 'generate',
    loadComponent: () => import('./pages/course-generator/course-generator.component').then(m => m.CourseGeneratorComponent)
  },
  {
    path: 'workflow',
    loadComponent: () => import('./pages/course-workflow/course-workflow.component').then(m => m.CourseWorkflowComponent)
  },
  {
    path: 'courses/:courseId/questions',
    loadComponent: () => import('./pages/question-editor/question-editor.component').then(m => m.QuestionEditorComponent)
  },
  {
    path: 'courses/:courseId/lessons',
    loadComponent: () => import('./pages/lesson-editor/lesson-editor.component').then(m => m.LessonEditorComponent)
  }
];
