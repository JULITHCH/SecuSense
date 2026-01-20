import { Pipe, PipeTransform } from '@angular/core';
import Prism from 'prismjs';

// Import additional languages
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-typescript';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-bash';
import 'prismjs/components/prism-sql';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-yaml';
import 'prismjs/components/prism-markup';
import 'prismjs/components/prism-css';

@Pipe({
  name: 'highlightCode',
  standalone: true
})
export class HighlightCodePipe implements PipeTransform {
  transform(html: string): string {
    if (!html) return '';

    // Parse the HTML and find code blocks to highlight
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = html;

    // Find all <code> elements and highlight them
    const codeElements = tempDiv.querySelectorAll('code');
    codeElements.forEach(codeEl => {
      const code = codeEl.textContent || '';
      const language = this.detectLanguage(code, codeEl.className);

      try {
        const grammar = Prism.languages[language] || Prism.languages['javascript'];
        const highlighted = Prism.highlight(code, grammar, language);
        codeEl.innerHTML = highlighted;
        codeEl.classList.add(`language-${language}`);
      } catch (e) {
        // If highlighting fails, leave original content
        console.warn('Code highlighting failed:', e);
      }
    });

    // Also handle <pre><code> blocks
    const preCodeElements = tempDiv.querySelectorAll('pre code');
    preCodeElements.forEach(codeEl => {
      const pre = codeEl.parentElement;
      if (pre) {
        pre.classList.add('code-block');
      }
    });

    return tempDiv.innerHTML;
  }

  private detectLanguage(code: string, className: string): string {
    // Check if language is specified in class
    const classMatch = className.match(/language-(\w+)/);
    if (classMatch) {
      return classMatch[1];
    }

    // Auto-detect based on content patterns
    if (code.includes('def ') || code.includes('import ') && code.includes(':')) {
      return 'python';
    }
    if (code.includes('function ') || code.includes('const ') || code.includes('let ')) {
      return 'javascript';
    }
    if (code.includes('interface ') || code.includes(': string') || code.includes(': number')) {
      return 'typescript';
    }
    if (code.includes('SELECT ') || code.includes('INSERT ') || code.includes('CREATE TABLE')) {
      return 'sql';
    }
    if (code.includes('#!/bin/bash') || code.includes('sudo ') || code.startsWith('$')) {
      return 'bash';
    }
    if (code.trim().startsWith('{') || code.trim().startsWith('[')) {
      try {
        JSON.parse(code);
        return 'json';
      } catch {
        // Not JSON
      }
    }
    if (code.includes('<html') || code.includes('<div') || code.includes('</')) {
      return 'markup';
    }

    // Default to javascript
    return 'javascript';
  }
}
