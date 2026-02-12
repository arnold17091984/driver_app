import { create } from 'zustand';
import { en, loadTranslations } from '../i18n';
import type { Locale, Translations, TranslationKey } from '../i18n';

interface I18nState {
  locale: Locale;
  translations: Translations;
  setLocale: (locale: Locale) => void;
  t: (key: TranslationKey, params?: Record<string, string | number>) => string;
}

function detectLocale(): Locale {
  const stored = localStorage.getItem('fleet_locale') as Locale | null;
  if (stored && ['en', 'ja', 'ko', 'zh'].includes(stored)) return stored;
  const lang = navigator.language.split('-')[0];
  if (['ja', 'ko', 'zh'].includes(lang)) return lang as Locale;
  return 'en';
}

function resolve(translations: Translations, key: string): string {
  const parts = key.split('.');
  let cur: unknown = translations;
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in cur) {
      cur = (cur as Record<string, unknown>)[p];
    } else {
      return key;
    }
  }
  return typeof cur === 'string' ? cur : key;
}

function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (_, k) =>
    params[k] !== undefined ? String(params[k]) : `{${k}}`
  );
}

const initialLocale = detectLocale();

export const useI18nStore = create<I18nState>((set, get) => {
  // Load non-en locale asynchronously
  if (initialLocale !== 'en') {
    loadTranslations(initialLocale).then((translations) => {
      set({ translations, locale: initialLocale });
    });
  }

  return {
    locale: initialLocale,
    translations: en,

    setLocale: (locale: Locale) => {
      localStorage.setItem('fleet_locale', locale);
      loadTranslations(locale).then((translations) => {
        set({ locale, translations });
      });
    },

    t: (key: TranslationKey, params?: Record<string, string | number>) => {
      const template = resolve(get().translations, key);
      return interpolate(template, params);
    },
  };
});
