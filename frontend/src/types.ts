export interface User {
  id: number;
  username: string;
  firstName: string;
  email: string;
  lastName: string;
  isVerified: boolean;
}

export interface Score {
  id: number;
  wordsPerMinute: number;
  wordsTyped: number;
  timeElapsed: number;
  accuracy: number;
  numberErrors: number;
  errors: { [error: string]: number };
  userId: number;
}

export type LanguageCode = (typeof languageCodes)[number];
export type LanguageName = (typeof languageNames)[number];

export interface Text {
  id: number;
  language: LanguageCode;
  text: string;
  punctuation: boolean;
  specialCharacters: number;
  numbers: number;
}

export interface Room {
  id: number;
}
