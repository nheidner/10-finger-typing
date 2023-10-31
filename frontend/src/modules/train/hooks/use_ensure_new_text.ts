import { useEffect } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createNewText,
  getAuthenticatedUser,
  getNewTextByUserid,
} from "@/utils/queries";
import { Text, LanguageCode } from "@/types";
import { useRouter } from "next/router";
import {
  numeralOptions,
  specialCharactersOptions,
} from "@/modules/train/constants";

const getRandomNumberBetween = (min: number, max: number) => {
  return Math.floor(Math.random() * (max - min + 1)) + min;
};

export const useEnsureTextData = ({
  specialCharacters,
  numerals,
  usePunctuation,
  language,
}: {
  specialCharacters: string;
  numerals: string;
  usePunctuation: boolean;
  language: LanguageCode;
}): { text?: Text; isLoading: boolean } => {
  const specialCharactersGte = specialCharactersOptions[
    specialCharacters
  ][0] as number;
  const specialCharactersLte = specialCharactersOptions[
    specialCharacters
  ][1] as number;
  const numbersGte = numeralOptions[numerals][0] as number;
  const numbersLte = numeralOptions[numerals][1] as number;

  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: getAuthenticatedUser,
    retry: false,
  });

  const { data: textData, isLoading: textIsLoading } = useQuery(
    [
      "text",
      specialCharactersGte,
      specialCharactersLte,
      numbersGte,
      numbersLte,
      usePunctuation,
      language,
    ],
    () =>
      getNewTextByUserid(authenticatedUserData?.id as number, {
        query: {
          specialCharactersGte,
          specialCharactersLte,
          numbersGte,
          numbersLte,
          punctuation: usePunctuation,
          language,
        },
      }),
    {
      enabled: !!authenticatedUserData?.id,
      onSuccess(data) {
        queryClient.setQueryData(["texts", data?.id], () => data);
        router.push({
          pathname: router.pathname,
          query: { ...router.query, textId: data?.id },
        });
      },
    }
  );

  const {
    mutate: mutateText,
    data: newTextData,
    isLoading: newTextIsLoading,
  } = useMutation({
    mutationKey: ["create text"],
    mutationFn: createNewText,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["text"] });
    },
  });

  useEffect(() => {
    if (textData !== null) {
      return;
    }

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
  }, [
    textData,
    language,
    specialCharactersLte,
    numbersLte,
    usePunctuation,
    specialCharactersGte,
    numbersGte,
    mutateText,
  ]);

  return {
    text: textData || newTextData,
    isLoading: (newTextIsLoading || textIsLoading) && !textData, // newTextIsLoading might still be true while textData is not null or undefined
  };
};
