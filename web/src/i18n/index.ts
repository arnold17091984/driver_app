import en from './en';
import type { Translations, Locale } from './types';

export type { Translations, Locale, TranslationKey } from './types';

const localeModules: Record<Locale, () => Promise<{ default: Translations }>> = {
  en: () => Promise.resolve({ default: en }),
  ja: () => import('./ja'),
  ko: () => import('./ko'),
  zh: () => import('./zh'),
};

const cache: Partial<Record<Locale, Translations>> = { en };

export async function loadTranslations(locale: Locale): Promise<Translations> {
  if (cache[locale]) return cache[locale]!;
  const mod = await localeModules[locale]();
  cache[locale] = mod.default;
  return mod.default;
}

export { en };
