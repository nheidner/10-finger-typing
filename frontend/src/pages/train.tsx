import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";
import { useState } from "react";
import { LanguageName } from "@/types";
import { Content } from "@/modules/train/components/Content";
import { useEnsureTextData } from "@/modules/train/hooks/use_ensure_new_text";
import { InviteModal } from "@/modules/train/components/invite_modal";
import {
  languageOptions,
  numeralOptions,
  specialCharactersOptions,
} from "@/modules/train/constants";
import { TextConfigOptions } from "@/components/TextConfigOptions";

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const [newRoomModalIsOpen, setNewRoomModalOpen] = useState(false);
  const userData = {};

  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [languageName, setLanguageName] = useState(
    Object.keys(languageOptions)[0] as LanguageName
  );
  const lang = languageOptions[languageName];

  const { text: textData, isLoading: textIsLoading } = useEnsureTextData({
    specialCharacters,
    numerals,
    usePunctuation,
    language: lang,
  });

  return (
    <>
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
          className="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
          onClick={() => setNewRoomModalOpen(true)}
        >
          Invite
        </button>
      </section>
      <InviteModal
        isOpen={newRoomModalIsOpen}
        setIsOpen={setNewRoomModalOpen}
      />
      <Content
        isActive={true}
        isLoading={textIsLoading}
        text={textData || null}
        userData={userData}
        onType={() => {}}
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
