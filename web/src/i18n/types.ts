import type en from './en';

/** Widen literal string types so translation files can use any string value. */
type DeepStringify<T> = {
  [K in keyof T]: T[K] extends Record<string, unknown> ? DeepStringify<T[K]> : string;
};

export type Translations = DeepStringify<typeof en>;

export type Locale = 'en' | 'ja' | 'ko' | 'zh';

type FlattenKeys<T, Prefix extends string = ''> = {
  [K in keyof T]: T[K] extends Record<string, unknown>
    ? FlattenKeys<T[K], `${Prefix}${K & string}.`>
    : `${Prefix}${K & string}`;
}[keyof T];

export type TranslationKey = FlattenKeys<Translations>;
