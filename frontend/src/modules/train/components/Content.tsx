import { ChangeEvent, FC, useEffect, useRef, useState } from "react";
import { Text, User } from "@/types";
import classNames from "classnames";
import { UserData } from "../types";

type LetterType = "correct" | "incorrect" | "notTyped";

export const Content: FC<{
  text: Text | null;
  isLoading: boolean;
  onType?: (cursor: number) => void;
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

    if (onType) {
      onType(cursorIndex + 1);
    }
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
