import {
  createNewText,
  getAuthenticatedUser,
  getNewTextByUserid,
} from "@/utils/queries";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { NextPage } from "next";
import { ChangeEvent, FC, useEffect, useState } from "react";
import { Switch as HeadlessUiSwitch } from "@headlessui/react";
import classNames from "classnames";
import { Text, TypingLanguage } from "@/types";

const specialCharactersOptions = {
  "0-5": "0-5",
  "5-10": "5-10",
  "10-15": "10-15",
  "15-20": "15-20",
};

const numeralOptions = {
  "0-5": "0-5",
  "5-10": "5-10",
  "10-15": "10-15",
  "15-20": "15-20",
};

const languageOptions: { [key in TypingLanguage]: string } = {
  de: "German",
  en: "English",
  fr: "French",
};

const Switch: FC<{
  item: string;
  label: string;
  enabled: boolean;
  handleChange: () => void;
}> = ({ item, label, enabled, handleChange }) => {
  return (
    <div className="flex items-center flex-col">
      <label
        htmlFor={item}
        className="block text-sm font-medium leading-6 text-gray-900 mb-2"
      >
        {label}
      </label>
      <HeadlessUiSwitch
        checked={enabled}
        onChange={handleChange}
        className={classNames(
          enabled ? "bg-indigo-600" : "bg-gray-200",
          "relative inline-flex h-9 w-14 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2"
        )}
      >
        <span className="sr-only">Use setting</span>
        <span
          aria-hidden="true"
          className={classNames(
            enabled ? "translate-x-5" : "translate-x-0",
            "pointer-events-none inline-block h-8 w-8 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
          )}
        />
      </HeadlessUiSwitch>
    </div>
  );
};

const Toggle: FC<{
  item: string;
  label: string;
  options: { [key: string]: string };
  selectedValue: string;
  handleChange: (e: ChangeEvent<HTMLSelectElement>) => void;
}> = ({ item, label, options, selectedValue, handleChange }) => {
  return (
    <div className="flex items-center flex-col">
      <label
        htmlFor={item}
        className="block text-sm font-medium leading-6 text-gray-900"
      >
        {label}
      </label>
      <select
        id={item}
        name={item}
        className="mt-2 block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
        defaultValue={Object.values(options)[0]}
        onChange={handleChange}
        value={selectedValue}
      >
        {Object.entries(options).map(([key, value]) => (
          <option key={key} value={key}>
            {value}
          </option>
        ))}
      </select>
    </div>
  );
};

const Content: FC<{ text: Text | null; isLoading: boolean }> = ({
  text,
  isLoading,
}) => {
  if (isLoading) {
    return <section>Loading...</section>;
  }

  return <section>{JSON.stringify(text)}</section>;
};

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const queryClient = useQueryClient();

  const { data: authenticatedUserData } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: () => getAuthenticatedUser(),
    retry: false,
  });

  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [language, setLanguage] = useState(
    Object.keys(languageOptions)[0] as TypingLanguage
  );

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

  const specialCharactersGte = parseInt(specialCharacters.split("-")[0], 10);
  const specialCharactersLte = parseInt(specialCharacters.split("-")[1], 10);
  const numbersGte = parseInt(numerals.split("-")[0], 10);
  const numbersLte = parseInt(numerals.split("-")[1], 10);

  const { data: textData, isLoading: textIsLoading } = useQuery(
    ["text", specialCharacters, numerals, usePunctuation, language],
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
    { enabled: !!authenticatedUserData?.id }
  );

  const {
    mutate,
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
    if (textData === null) {
      const specialCharacters = specialCharactersLte - 2;
      const numbers = numbersLte - 2;

      mutate({
        query: {
          specialCharacters,
          numbers,
          punctuation: usePunctuation,
          language,
        },
      });
    }
  }, [textData]);

  return (
    <>
      <section className="flex gap-10 justify-center">
        <Toggle
          item="specialCharacters"
          label="Special Characters"
          options={specialCharactersOptions}
          selectedValue={specialCharacters}
          handleChange={handleSpecialCharactersChange}
        />
        <Toggle
          item="numerals"
          label="Number of Numerals"
          options={numeralOptions}
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
          options={languageOptions}
          selectedValue={language}
          handleChange={handleLanguageChange}
        />
      </section>
      <Content
        isLoading={newTextIsLoading || textIsLoading}
        text={textData || null}
      />
    </>
  );
};

TrainPage.getInitialProps = async (ctx) => {
  const queryClient = new QueryClient();

  return {
    dehydratedState: dehydrate(queryClient),
  };
};

export default TrainPage;
