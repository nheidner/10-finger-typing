import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";
import { ChangeEvent, useEffect, useRef, useState } from "react";
import { TypingLanguage, User } from "@/types";
import { Toggle } from "@/components/train/Toggle";
import { Switch } from "@/components/train/Switch";
import { getWsUrl } from "@/utils/get_api_url";
import { Content } from "@/components/train/Content";
import { useEnsureTextData } from "@/hooks/use_ensure_new_text";

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

type LetterType = "correct" | "incorrect" | "notTyped";

type UserData = {
  user: User;
  cursor: number;
};

const TrainPage: NextPage<{
  dehydratedState: DehydratedState;
}> = () => {
  const webSocketRef = useRef<WebSocket | null>(null);

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

  useEffect(() => {
    const apiUrl = getWsUrl();

    // runs after there is textData
    if (textData?.id) {
      const websocketUrl = `${apiUrl}/texts/${textData.id}/rooms/fa1c3fc5-a018-4d92-9c53-47ade506e1f0/ws`;
      webSocketRef.current = new WebSocket(websocketUrl);

      webSocketRef.current.onopen = () => {
        const message = {
          type: "user_added",
        };
        webSocketRef.current?.send(JSON.stringify(message));
      };

      webSocketRef.current.onmessage = (e) => {
        type Message = {
          user: User;
          type: "user_added" | "cursor";
          payload: { cursor: number } | null;
        };

        const message = JSON.parse(e.data) as Message;

        switch (message.type) {
          case "user_added":
            setUserData((userData) => {
              return {
                ...userData,
                [message.user.id]: { user: message.user, cursor: 0 },
              };
            });
            break;
          case "cursor":
            setUserData((userData) => {
              return {
                ...userData,
                [message.user.id]: {
                  ...userData[message.user.id],
                  cursor: message.payload?.cursor || 0,
                },
              };
            });
          default:
            break;
        }

        console.log("message received: >>", JSON.parse(e.data));
      };

      webSocketRef.current.onclose = (e) => {
        // console.log("connection closed: >>", e);
      };
    }
    return () => {
      webSocketRef.current?.close();
    };
  }, [textData?.id]);

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
