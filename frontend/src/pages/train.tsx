import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";
import { ChangeEvent, useRef, useState } from "react";
import { TypingLanguage, User } from "@/types";
import { Toggle } from "@/modules/train/components/Toggle";
import { Switch } from "@/modules/train/components/Switch";
import { Content } from "@/modules/train/components/Content";
import { useEnsureTextData } from "@/modules/train/hooks/use_ensure_new_text";
import { useConnectToRoom } from "@/modules/train/hooks/use_connect_to_room";
import { UserData } from "@/modules/train/types";

const specialCharactersOptions = {
  "0-4": "0-4",
  "5-9": "5-9",
  "10-14": "10-14",
  "15-19": "15-19",
};

const numeralOptions = {
  "0-4": "0-4",
  "5-9": "5-9",
  "10-14": "10-14",
  "15-19": "15-19",
};

const languageOptions: { [key in TypingLanguage]: string } = {
  de: "German",
  en: "English",
  fr: "French",
};

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [language, setLanguage] = useState(
    Object.keys(languageOptions)[0] as TypingLanguage
  );
  const [userData, setUserData] = useState<{ [userId: number]: UserData }>({});

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

  const { text: textData, isLoading: textIsLoading } = useEnsureTextData({
    specialCharactersGte,
    specialCharactersLte,
    numbersGte,
    numbersLte,
    usePunctuation,
    language,
  });

  const roomId = "b4df1403-1599-48f1-9ea2-36dc4d97cfc0";

  const webSocketRef = useConnectToRoom(roomId, setUserData, textData);

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
        isLoading={textIsLoading}
        text={textData || null}
        userData={userData}
        onType={(cursor: number) => {
          const message = {
            type: "cursor",
            payload: {
              cursor,
            },
          };
          webSocketRef.current?.send(JSON.stringify(message));
        }}
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
