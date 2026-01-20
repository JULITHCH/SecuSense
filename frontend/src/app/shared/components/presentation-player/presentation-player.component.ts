import { Component, input, signal, effect, ElementRef, ViewChild, AfterViewInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ButtonModule } from 'primeng/button';
import { ProgressBarModule } from 'primeng/progressbar';
import { TooltipModule } from 'primeng/tooltip';
import { HighlightCodePipe } from '../../pipes/highlight-code.pipe';

export interface PresentationSlide {
  title: string;
  content: string;
  script: string;
  audioUrl: string;
  imageUrl?: string;
  imageAlt?: string;
}

@Component({
  selector: 'app-presentation-player',
  standalone: true,
  imports: [CommonModule, ButtonModule, ProgressBarModule, TooltipModule, HighlightCodePipe],
  template: `
    <div class="presentation-container">
      <!-- Slide Display -->
      <div class="slide-wrapper">
        <div class="slide" [class.transitioning]="transitioning()">
          @if (currentSlide()) {
            <h2 class="slide-title">{{ currentSlide()!.title }}</h2>

            <!-- Stock Image (if available) -->
            @if (currentSlide()!.imageUrl) {
              <div class="slide-image">
                <img
                  [src]="currentSlide()!.imageUrl"
                  [alt]="currentSlide()!.imageAlt || currentSlide()!.title"
                  loading="lazy" />
              </div>
            }

            <div class="slide-content" [innerHTML]="currentSlide()!.content | highlightCode"></div>
          }
        </div>

        <!-- Script Panel (collapsible) -->
        @if (showScript() && currentSlide()?.script) {
          <div class="script-panel" [class.expanded]="showScript()">
            <div class="script-header">
              <i class="pi pi-file-edit"></i>
              <span>Narrator Script</span>
            </div>
            <div class="script-content">
              {{ currentSlide()!.script }}
            </div>
          </div>
        }
      </div>

      <!-- Progress Bar -->
      <div class="progress-section">
        <p-progressBar
          [value]="progressPercent()"
          [showValue]="false"
          styleClass="presentation-progress">
        </p-progressBar>
        <span class="slide-counter">{{ currentIndex() + 1 }} / {{ slides().length }}</span>
      </div>

      <!-- Controls -->
      <div class="controls">
        <p-button
          icon="pi pi-step-backward"
          [outlined]="true"
          [disabled]="currentIndex() === 0"
          (onClick)="previousSlide()"
          pTooltip="Previous slide">
        </p-button>

        <p-button
          [icon]="isPlaying() ? 'pi pi-pause' : 'pi pi-play'"
          (onClick)="togglePlayPause()"
          pTooltip="Play/Pause audio">
        </p-button>

        <p-button
          icon="pi pi-step-forward"
          [outlined]="true"
          [disabled]="currentIndex() === slides().length - 1"
          (onClick)="nextSlide()"
          pTooltip="Next slide">
        </p-button>

        <p-button
          icon="pi pi-volume-up"
          [outlined]="true"
          [disabled]="!currentSlide()?.audioUrl"
          (onClick)="replayAudio()"
          pTooltip="Replay audio">
        </p-button>

        <div class="controls-separator"></div>

        <p-button
          [icon]="showScript() ? 'pi pi-eye-slash' : 'pi pi-file-edit'"
          [outlined]="true"
          (onClick)="toggleScript()"
          [pTooltip]="showScript() ? 'Hide script' : 'Show script'">
        </p-button>
      </div>

      <!-- Hidden Audio Element -->
      <audio #audioPlayer (ended)="onAudioEnded()" (play)="onAudioPlay()" (pause)="onAudioPause()"></audio>
    </div>
  `,
  styles: [`
    .presentation-container {
      display: flex;
      flex-direction: column;
      height: 100%;
      min-height: 500px;
      background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
      border-radius: 12px;
      overflow: hidden;
    }

    .slide-wrapper {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: 2rem;
      overflow: auto;
    }

    .slide {
      flex: 1;
      width: 100%;
      max-width: 900px;
      margin: 0 auto;
      background: rgba(255, 255, 255, 0.05);
      border-radius: 12px;
      padding: 3rem;
      color: #fff;
      transition: opacity 0.3s ease, transform 0.3s ease;
    }

    .slide.transitioning {
      opacity: 0.5;
      transform: scale(0.98);
    }

    .slide-title {
      font-size: 2rem;
      font-weight: 600;
      margin-bottom: 2rem;
      color: #4fc3f7;
      text-align: center;
    }

    .slide-image {
      margin-bottom: 1.5rem;
      text-align: center;
    }

    .slide-image img {
      max-width: 100%;
      max-height: 250px;
      border-radius: 8px;
      object-fit: cover;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    }

    .slide-content {
      font-size: 1.25rem;
      line-height: 1.8;
    }

    .slide-content :deep(ul) {
      list-style: none;
      padding: 0;
    }

    .slide-content :deep(li) {
      padding: 0.75rem 0;
      padding-left: 2rem;
      position: relative;
    }

    .slide-content :deep(li)::before {
      content: "\\2022";
      color: #4fc3f7;
      font-size: 1.5rem;
      position: absolute;
      left: 0;
      top: 0.5rem;
    }

    .slide-content :deep(strong) {
      color: #81d4fa;
    }

    /* Basic code styling */
    .slide-content :deep(code) {
      background: rgba(0, 0, 0, 0.4);
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-family: 'Fira Code', 'Consolas', monospace;
      font-size: 0.9em;
    }

    /* Code block styling */
    .slide-content :deep(pre) {
      background: rgba(0, 0, 0, 0.5);
      padding: 1rem;
      border-radius: 8px;
      overflow-x: auto;
      margin: 1rem 0;
    }

    .slide-content :deep(pre code) {
      background: transparent;
      padding: 0;
      display: block;
      line-height: 1.6;
    }

    /* Prism.js syntax highlighting colors (dark theme) */
    .slide-content :deep(.token.comment),
    .slide-content :deep(.token.prolog),
    .slide-content :deep(.token.doctype),
    .slide-content :deep(.token.cdata) {
      color: #6a9955;
    }

    .slide-content :deep(.token.punctuation) {
      color: #d4d4d4;
    }

    .slide-content :deep(.token.property),
    .slide-content :deep(.token.tag),
    .slide-content :deep(.token.boolean),
    .slide-content :deep(.token.number),
    .slide-content :deep(.token.constant),
    .slide-content :deep(.token.symbol) {
      color: #b5cea8;
    }

    .slide-content :deep(.token.selector),
    .slide-content :deep(.token.attr-name),
    .slide-content :deep(.token.string),
    .slide-content :deep(.token.char),
    .slide-content :deep(.token.builtin) {
      color: #ce9178;
    }

    .slide-content :deep(.token.operator),
    .slide-content :deep(.token.entity),
    .slide-content :deep(.token.url),
    .slide-content :deep(.token.variable) {
      color: #d4d4d4;
    }

    .slide-content :deep(.token.atrule),
    .slide-content :deep(.token.attr-value),
    .slide-content :deep(.token.function),
    .slide-content :deep(.token.class-name) {
      color: #dcdcaa;
    }

    .slide-content :deep(.token.keyword) {
      color: #569cd6;
    }

    .slide-content :deep(.token.regex),
    .slide-content :deep(.token.important) {
      color: #d16969;
    }

    /* Script Panel */
    .script-panel {
      margin-top: 1rem;
      background: rgba(0, 0, 0, 0.3);
      border-radius: 8px;
      border: 1px solid rgba(79, 195, 247, 0.3);
      overflow: hidden;
      animation: slideDown 0.3s ease;
    }

    @keyframes slideDown {
      from {
        opacity: 0;
        transform: translateY(-10px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    .script-header {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.75rem 1rem;
      background: rgba(79, 195, 247, 0.1);
      color: #4fc3f7;
      font-weight: 500;
      font-size: 0.875rem;
    }

    .script-content {
      padding: 1rem;
      color: rgba(255, 255, 255, 0.9);
      font-size: 1rem;
      line-height: 1.7;
      max-height: 200px;
      overflow-y: auto;
    }

    .progress-section {
      padding: 1rem 2rem;
      display: flex;
      align-items: center;
      gap: 1rem;
      background: rgba(0, 0, 0, 0.2);
    }

    .progress-section :deep(.presentation-progress) {
      flex: 1;
      height: 6px;
    }

    .slide-counter {
      color: rgba(255, 255, 255, 0.7);
      font-size: 0.875rem;
      min-width: 60px;
      text-align: right;
    }

    .controls {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 1rem;
      padding: 1.5rem;
      background: rgba(0, 0, 0, 0.3);
    }

    .controls-separator {
      width: 1px;
      height: 24px;
      background: rgba(255, 255, 255, 0.2);
      margin: 0 0.5rem;
    }

    .controls :deep(.p-button) {
      width: 48px;
      height: 48px;
    }
  `]
})
export class PresentationPlayerComponent implements AfterViewInit, OnDestroy {
  slides = input.required<PresentationSlide[]>();
  autoPlay = input<boolean>(true);

  @ViewChild('audioPlayer') audioPlayer!: ElementRef<HTMLAudioElement>;

  currentIndex = signal(0);
  isPlaying = signal(false);
  transitioning = signal(false);
  showScript = signal(false);

  currentSlide = signal<PresentationSlide | null>(null);

  progressPercent = signal(0);

  private audioEndedHandler = () => this.onAudioEnded();

  constructor() {
    effect(() => {
      const slides = this.slides();
      const index = this.currentIndex();
      if (slides && slides.length > 0 && index >= 0 && index < slides.length) {
        this.currentSlide.set(slides[index]);
        this.progressPercent.set(((index + 1) / slides.length) * 100);
      }
    }, { allowSignalWrites: true });
  }

  ngAfterViewInit(): void {
    // Start playing first slide if autoPlay is enabled
    setTimeout(() => {
      if (this.autoPlay() && this.slides().length > 0) {
        this.playCurrentSlideAudio();
      }
    }, 500);
  }

  ngOnDestroy(): void {
    if (this.audioPlayer?.nativeElement) {
      this.audioPlayer.nativeElement.pause();
    }
  }

  toggleScript(): void {
    this.showScript.update(v => !v);
  }

  nextSlide(): void {
    const slides = this.slides();
    if (this.currentIndex() < slides.length - 1) {
      this.transitioning.set(true);
      setTimeout(() => {
        this.currentIndex.update(i => i + 1);
        this.transitioning.set(false);
        this.playCurrentSlideAudio();
      }, 300);
    }
  }

  previousSlide(): void {
    if (this.currentIndex() > 0) {
      this.transitioning.set(true);
      setTimeout(() => {
        this.currentIndex.update(i => i - 1);
        this.transitioning.set(false);
        this.playCurrentSlideAudio();
      }, 300);
    }
  }

  togglePlayPause(): void {
    const audio = this.audioPlayer?.nativeElement;
    if (!audio) return;

    if (this.isPlaying()) {
      audio.pause();
    } else {
      if (audio.src) {
        audio.play();
      } else {
        this.playCurrentSlideAudio();
      }
    }
  }

  replayAudio(): void {
    const audio = this.audioPlayer?.nativeElement;
    if (audio && audio.src) {
      audio.currentTime = 0;
      audio.play();
    }
  }

  playCurrentSlideAudio(): void {
    const slide = this.currentSlide();
    const audio = this.audioPlayer?.nativeElement;

    if (!audio || !slide?.audioUrl) {
      return;
    }

    audio.src = slide.audioUrl;
    audio.play().catch(err => {
      console.warn('Audio playback failed:', err);
    });
  }

  onAudioEnded(): void {
    this.isPlaying.set(false);
    // Auto-advance to next slide if not on last slide
    if (this.autoPlay() && this.currentIndex() < this.slides().length - 1) {
      setTimeout(() => this.nextSlide(), 1000);
    }
  }

  onAudioPlay(): void {
    this.isPlaying.set(true);
  }

  onAudioPause(): void {
    this.isPlaying.set(false);
  }
}
