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
import { FetchError } from "@/utils/fetch";
import { getRandomNumberBetween } from "@/utils/random";

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

  const {
    data: textData,
    isLoading: textIsLoading,
    status,
  } = useQuery(
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
    {
      enabled: !!authenticatedUserData?.id,
      retry: false,
      onSuccess(data) {
        queryClient.setQueryData(["texts", data?.id], () => data);
        router.push({
          pathname: router.pathname,
          query: { ...router.query, textId: data?.id || "" },
        });
      },
      onError(err) {
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

  return {
    text: textData || newTextData,
    isLoading: (newTextIsLoading || textIsLoading) && !textData, // newTextIsLoading might still be true while textData is not null or undefined
  };
};
