import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";
import { ChangeEvent, useState } from "react";
import { TypingLanguage } from "@/types";
import { Toggle } from "@/modules/train/components/Toggle";
import { Switch } from "@/modules/train/components/Switch";
import { Content } from "@/modules/train/components/Content";
import { useEnsureTextData } from "@/modules/train/hooks/use_ensure_new_text";
import { useConnectToRoom } from "@/modules/train/hooks/use_connect_to_room";
import { UserData } from "@/modules/train/types";
import { InviteModal } from "@/modules/train/components/InviteModal";

const specialCharactersOptions: { [value: string]: number[] | TypingLanguage } =
  {
    "0-4": [0, 4],
    "5-9": [5, 9],
    "10-14": [10, 14],
    "15-19": [15, 19],
  };

const numeralOptions: { [value: string]: number[] | TypingLanguage } = {
  "0-4": [0, 4],
  "5-9": [5, 9],
  "10-14": [10, 14],
  "15-19": [15, 19],
};

const languageOptions: { [value: string]: number[] | TypingLanguage } = {
  English: "en",
  German: "de",
  French: "fr",
};

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const [newRoomUsers, setNewUsers] = useState<string[]>([]);
  const [newRoomModalIsOpen, setNewRoomModalOpen] = useState(false);
  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [language, setLanguage] = useState(Object.keys(languageOptions)[0]);
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
    setLanguage(e.target.value);
  };
  const handleSelectNewRoomUser = (user: string) => {
    setNewUsers((users) => users.concat(user));
  };

  const specialCharactersGte = specialCharactersOptions[
    specialCharacters
  ][0] as number;
  const specialCharactersLte = specialCharactersOptions[
    specialCharacters
  ][1] as number;
  const numbersGte = numeralOptions[numerals][0] as number;
  const numbersLte = numeralOptions[numerals][1] as number;
  const lang = languageOptions[language] as TypingLanguage;

  const { text: textData, isLoading: textIsLoading } = useEnsureTextData({
    specialCharactersGte,
    specialCharactersLte,
    numbersGte,
    numbersLte,
    usePunctuation,
    language: lang,
  });

  const roomId = "b4df1403-1599-48f1-9ea2-36dc4d97cfc0";

  const webSocketRef = useConnectToRoom(roomId, setUserData, textData);

  return (
    <>
      <section className="flex gap-10 justify-center items-center">
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
        <button
          type="button"
          className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
          onClick={() => setNewRoomModalOpen(true)}
        >
          Invite
        </button>
      </section>
      <InviteModal
        isOpen={newRoomModalIsOpen}
        setOpen={setNewRoomModalOpen}
        handleSelectNewRoomUser={handleSelectNewRoomUser}
        newRoomUsers={newRoomUsers}
      />
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
