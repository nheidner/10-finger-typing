import { ChangeEvent, FC } from "react";
import { TypingLanguage } from "@/types";
import { Toggle } from "@/modules/train/components/Toggle";
import { Switch } from "@/modules/train/components/Switch";
import {
  languageOptions,
  numeralOptions,
  specialCharactersOptions,
} from "@/modules/train/constants";

export const TextConfigOptions: FC<{
  specialCharacters: string;
  setSpecialCharacters: (option: string) => void;
  numerals: string;
  setNumerals: (option: string) => void;
  usePunctuation: boolean;
  setUsePunctuation: (option: boolean) => void;
  language: string;
  setLanguage: (option: string) => void;
}> = ({
  specialCharacters,
  setSpecialCharacters,
  numerals,
  setNumerals,
  usePunctuation,
  setUsePunctuation,
  language,
  setLanguage,
}) => {
  const handleSpecialCharactersChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setSpecialCharacters(e.target.value);
  };
  const handleNumeralsChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setNumerals(e.target.value);
  };
  const handlePunctuationChange = () => {
    setUsePunctuation(!usePunctuation);
  };
  const handleLanguageChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setLanguage(e.target.value as TypingLanguage);
  };

  return (
    <>
      <Toggle
        item="specialCharacters"
        label="Special Characters"
        options={Object.keys(specialCharactersOptions)}
        selectedValue={specialCharacters}
        handleChange={handleSpecialCharactersChange}
      />
      <Toggle
        item="numerals"
        label="Number of Numerals"
        options={Object.keys(numeralOptions)}
        selectedValue={numerals}
        handleChange={handleNumeralsChange}
      />
      <Switch
        item="usePunctuation"
        label="Use Punctuation"
        enabled={usePunctuation}
        handleChange={handlePunctuationChange}
      />
      <Toggle
        item="languages"
        label="Languages"
        options={Object.keys(languageOptions)}
        selectedValue={language}
        handleChange={handleLanguageChange}
      />
    </>
  );
};
