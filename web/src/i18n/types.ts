import type en from './en';

export type Translations = typeof en;

export type Locale = 'en' | 'ja' | 'ko' | 'zh';

type FlattenKeys<T, Prefix extends string = ''> = {
  [K in keyof T]: T[K] extends Record<string, unknown>
    ? FlattenKeys<T[K], `${Prefix}${K & string}.`>
    : `${Prefix}${K & string}`;
}[keyof T];

export type TranslationKey = FlattenKeys<Translations>;
