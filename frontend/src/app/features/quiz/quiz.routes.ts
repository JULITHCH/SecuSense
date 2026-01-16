import { Routes } from '@angular/router';

export const QUIZ_ROUTES: Routes = [
  {
    path: ':courseId',
    loadComponent: () => import('./pages/quiz-container/quiz-container.component').then(m => m.QuizContainerComponent)
  },
  {
    path: 'results/:attemptId',
    loadComponent: () => import('./pages/results/results.component').then(m => m.ResultsComponent)
  }
];
