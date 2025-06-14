import { marked } from 'marked';
import hljs from 'highlight.js';

// Configure marked options
marked.setOptions({
  breaks: true,
  gfm: true,
  highlight: (code: string, lang: string) => {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value;
      } catch (err) {
        console.error('Error highlighting code:', err);
      }
    }
    return code;
  }
});

// Sanitize HTML to prevent XSS
function sanitizeHtml(html: string): string {
  const div = document.createElement('div');
  div.innerHTML = html;
  
  // Remove potentially dangerous attributes
  const elements = div.getElementsByTagName('*');
  for (let i = 0; i < elements.length; i++) {
    const element = elements[i];
    const attributes = element.attributes;
    for (let j = attributes.length - 1; j >= 0; j--) {
      const attr = attributes[j];
      if (attr.name.startsWith('on') || attr.name === 'href' && attr.value.startsWith('javascript:')) {
        element.removeAttribute(attr.name);
      }
    }
  }
  
  return div.innerHTML;
}

export function renderMarkdown(text: string): string {
  const html = marked.parse(text) as string;
  return sanitizeHtml(html);
} 