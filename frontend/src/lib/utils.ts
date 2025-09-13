import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function detectLanguage(text: string): string {
  // Simple language detection based on character patterns
  const cyrillicPattern = /[\u0400-\u04FF]/;
  const chinesePattern = /[\u4e00-\u9fff]/;
  const arabicPattern = /[\u0600-\u06FF]/;
  
  if (cyrillicPattern.test(text)) return 'ru';
  if (chinesePattern.test(text)) return 'zh';
  if (arabicPattern.test(text)) return 'ar';
  
  return 'en';
}

export function detectRegion(lang: string): string {
  const regionMap: Record<string, string> = {
    'en': 'US',
    'ru': 'RU',
    'zh': 'CN',
    'ar': 'AE',
    'es': 'ES',
    'fr': 'FR',
    'de': 'DE',
    'it': 'IT',
    'ja': 'JP',
    'ko': 'KR'
  };
  
  return regionMap[lang] || 'US';
}

export function formatPrice(price: number, currency: string = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
  }).format(price);
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '...';
}