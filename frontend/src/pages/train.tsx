import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";
import { useState } from "react";
import { TypingLanguage } from "@/types";
import { Content } from "@/modules/train/components/Content";
import { useEnsureTextData } from "@/modules/train/hooks/use_ensure_new_text";
import { useConnectToRoom } from "@/modules/train/hooks/use_connect_to_room";
import { UserData } from "@/modules/train/types";
import { InviteModal } from "@/modules/train/components/invite_modal";
import {
  languageOptions,
  numeralOptions,
  specialCharactersOptions,
} from "@/modules/train/constants";
import { useRouter } from "next/router";
import { TextConfigOptions } from "@/components/TextConfigOptions";

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const [newRoomModalIsOpen, setNewRoomModalOpen] = useState(false);
  const [userData, setUserData] = useState<{ [userId: number]: UserData }>({});

  const [specialCharacters, setSpecialCharacters] = useState(
    Object.keys(specialCharactersOptions)[0]
  );
  const [numerals, setNumerals] = useState(Object.keys(numeralOptions)[0]);
  const [usePunctuation, setUsePunctuation] = useState(false);
  const [language, setLanguage] = useState(Object.keys(languageOptions)[0]);
  const lang = languageOptions[language] as TypingLanguage;

  const { text: textData, isLoading: textIsLoading } = useEnsureTextData({
    specialCharacters,
    numerals,
    usePunctuation,
    language: lang,
  });

  const router = useRouter();
  const { roomId } = router.query as {
    roomId?: string;
  };

  const webSocketRef = useConnectToRoom(setUserData, roomId, textData?.id);

  return (
    <>
      <section className="flex gap-10 justify-center items-center">
        <TextConfigOptions
          setLanguage={setLanguage}
          setNumerals={setNumerals}
          setSpecialCharacters={setSpecialCharacters}
          setUsePunctuation={setUsePunctuation}
          specialCharacters={specialCharacters}
          language={language}
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
      <InviteModal isOpen={newRoomModalIsOpen} setOpen={setNewRoomModalOpen} />
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
