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
import { ChangeEvent, FC, use, useEffect, useRef, useState } from "react";
import { Text, TypingLanguage, User } from "@/types";
import { Toggle } from "@/components/train/Toggle";
import { Switch } from "@/components/train/Switch";
import classNames from "classnames";

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

type LetterType = "correct" | "incorrect" | "notTyped";

const Content: FC<{
  text: Text | null;
  isLoading: boolean;
  onType: (cursor: number) => void;
  userData: { [userId: number]: UserData };
}> = ({ text, isLoading, onType, userData }) => {
  const [editedText, setEditedText] = useState<
    {
      char: string;
      type: LetterType;
    }[]
  >([]);

  const [input, setInput] = useState<string>("");
  const [cursorIndex, setCursorIndex] = useState<number>(0);

  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (!text) {
      return;
    }

    const newText = text.text.split("").map((char) => {
      return {
        char,
        type: "notTyped" as LetterType,
      };
    });

    setEditedText(newText);
  }, [text]);

  useEffect(() => {
    setInput("");
    setCursorIndex(0);
  }, [text]);

  const focusTextInput = () => {
    inputRef.current?.focus();
  };

  useEffect(() => {
    window.addEventListener("keydown", focusTextInput);
    return () => {
      window.removeEventListener("keydown", focusTextInput);
    };
  }, []);

  const handleInput = (e: ChangeEvent<HTMLInputElement>) => {
    const newChar = e.target.value.slice(-1);

    const isCorrectChar = newChar === editedText[cursorIndex].char;

    if (!isCorrectChar) {
      setEditedText((editedText) => {
        const newEditedText = [...editedText];
        newEditedText[cursorIndex].type = "incorrect";
        return newEditedText;
      });
      return;
    }

    if (editedText[cursorIndex].type !== "incorrect") {
      setEditedText((editedText) => {
        const newEditedText = [...editedText];
        newEditedText[cursorIndex].type = "correct";
        return newEditedText;
      });
    }

    onType(cursorIndex + 1);
    setInput((input) => input + newChar);
    setCursorIndex((cursorIndex) => cursorIndex + 1);
  };

  if (isLoading) {
    return <section>Loading...</section>;
  }

  if (!text) {
    return <section>No text found</section>;
  }

  return (
    <section className="relative">
      <input
        type="text"
        className="absolute opacity-0"
        onChange={handleInput}
        ref={inputRef}
        value={input}
      />
      <div className="font-courier" onClick={focusTextInput}>
        {editedText.map((char, index) => {
          const otherUserCursor = Object.entries(userData).reduce(
            (acc, [_, userData]) => {
              return userData.cursor === index || acc;
            },
            false
          );

          return (
            <span
              key={index}
              className={classNames(
                otherUserCursor &&
                  "before:content-[] border-b-4 border-gray-200",
                cursorIndex === index &&
                  "before:content-[] border-b-4 border-gray-400",
                char.type === "correct" && "text-green-700",
                char.type === "incorrect" && "text-red-700",
                "text-2xl"
              )}
            >
              {char.char}
            </span>
          );
        })}
      </div>
    </section>
  );
};

type UserData = {
  user: User;
  cursor: number;
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

  useEffect(() => {
    const apiUrl =
      typeof window === "undefined"
        ? "ws://server_dev:8080/api"
        : "ws://localhost:8080/api";

    webSocketRef.current = new WebSocket(apiUrl + "/ws");

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
      console.log("connection closed: >>", e);
    };

    return () => {
      webSocketRef.current?.close();
    };
  }, []);

  console.log("userData :>> ", userData);

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
