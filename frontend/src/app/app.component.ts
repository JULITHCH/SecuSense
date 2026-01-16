import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ToastModule } from 'primeng/toast';
import { MessageService } from 'primeng/api';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, ToastModule],
  providers: [MessageService],
  template: `
    <p-toast></p-toast>
    <router-outlet></router-outlet>
  `,
  styles: []
})
export class AppComponent {
  title = 'SecuSense Training Portal';
}
