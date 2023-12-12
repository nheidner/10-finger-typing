import { TextConfigOptions } from "@/components/TextConfigOptions";
import {
  languageOptions,
  numeralOptions,
  specialCharactersOptions,
} from "@/modules/train/constants";
import { GameStatus, LanguageName } from "@/types";
import { FC, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  createGame,
  createNewText,
  getAuthenticatedUser,
  getNewTextByUserid,
} from "@/utils/queries";
import { FetchError } from "@/utils/fetch";
import { getRandomNumberBetween } from "@/utils/random";
import { LoadingSpinner } from "@/components/LoadingSpinner";

export const SelectNewText: FC<{
  roomId: string;
}> = ({ roomId }) => {
  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [languageName, setLanguageName] = useState(
    Object.keys(languageOptions)[0] as LanguageName
  );

  const specialCharactersGte = specialCharactersOptions[
    specialCharacters
  ][0] as number;
  const specialCharactersLte = specialCharactersOptions[
    specialCharacters
  ][1] as number;
  const numbersGte = numeralOptions[numerals][0] as number;
  const numbersLte = numeralOptions[numerals][1] as number;
  const language = languageOptions[languageName];

  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const { mutate: createNewGame, isLoading: createNewGameIsLoading } =
    useMutation({
      mutationKey: ["create new game"],
      mutationFn: async (textId: string) =>
        createGame({ roomId, body: { textId } }),
    });

  const { mutate: mutateText, isLoading: createNewTextIsLoading } = useMutation(
    {
      mutationKey: ["create text"],
      mutationFn: createNewText,
      onSuccess: (data) => {
        createNewGame(data.id);
      },
    }
  );

  const { mutate: getNewText, isLoading: getNewTextIsLoading } = useMutation({
    mutationKey: [
      "text",
      specialCharactersGte,
      specialCharactersLte,
      numbersGte,
      numbersLte,
      usePunctuation,
      language,
    ],
    mutationFn: () =>
      getNewTextByUserid(authenticatedUserData?.id as string, {
        query: {
          specialCharactersGte,
          specialCharactersLte,
          numbersGte,
          numbersLte,
          punctuation: usePunctuation,
          language,
        },
      }),
    onSuccess: (data) => {
      createNewGame(data.id);
    },
    onError: (err: any) => {
      if (err instanceof FetchError) {
        const specialCharacters = getRandomNumberBetween(
          specialCharactersGte,
          specialCharactersLte
        );
        const numbers = getRandomNumberBetween(numbersGte, numbersLte);

        mutateText({
          query: {
            specialCharacters,
            numbers,
            punctuation: usePunctuation,
            language,
          },
        });
      }
    },
  });

  const handleCreateNewGame = async () => {
    getNewText();
  };

  return (
    <section className="flex gap-10 justify-center items-center">
      <TextConfigOptions
        setLanguage={setLanguageName}
        setNumerals={setNumerals}
        setSpecialCharacters={setSpecialCharacters}
        setUsePunctuation={setUsePunctuation}
        specialCharacters={specialCharacters}
        language={languageName}
        numerals={numerals}
        usePunctuation={usePunctuation}
      />
      <button
        type="button"
        onClick={handleCreateNewGame}
        className="inline-flex items-center gap-2 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:bg-slate-500 disabled:hover:bg-slate-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
      >
        <div>create new game</div>
        <LoadingSpinner
          isLoading={
            createNewGameIsLoading ||
            createNewTextIsLoading ||
            getNewTextIsLoading
          }
          isWhite
        />
      </button>
    </section>
  );
};
