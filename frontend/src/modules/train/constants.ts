import { TypingLanguage } from "@/types";

export const specialCharactersOptions: {
  [value: string]: number[] | TypingLanguage;
} = {
  "0-4": [0, 4],
  "5-9": [5, 9],
  "10-14": [10, 14],
  "15-19": [15, 19],
};

export const numeralOptions: { [value: string]: number[] | TypingLanguage } = {
  "0-4": [0, 4],
  "5-9": [5, 9],
  "10-14": [10, 14],
  "15-19": [15, 19],
};

export const languageOptions: { [value: string]: number[] | TypingLanguage } = {
  English: "en",
  German: "de",
  French: "fr",
};
