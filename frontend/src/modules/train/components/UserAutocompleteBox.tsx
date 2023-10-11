import { ChangeEvent, FC, useState } from "react";
import { Combobox } from "@headlessui/react";
import { useDebouncedUserSearchByUsernamePartial } from "../hooks/use_debounced_user_search_by_username_partial.ts";
import isEmail from "validator/lib/isEmail";
import { User } from "@/types";
import { AutocompleteOption } from "./AutocompleteOption";

export const UserAutocompleteBox: FC<{
  addNewRoomUser: (user: Partial<User>) => void;
  newRoomUsersDisplaySet: Set<string>;
}> = ({ addNewRoomUser, newRoomUsersDisplaySet }) => {
  const [input, setInput] = useState("");

  const { users, debouncedFetchUsers } =
    useDebouncedUserSearchByUsernamePartial(200);

  const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;

    setInput(value);
    debouncedFetchUsers(value);
  };

  const handleOnChange = (user: Partial<User>) => {
    setInput("");
    addNewRoomUser(user);
  };

  const isValidEmail = isEmail(input);
  const suggestedUsers = users?.filter(
    (suggestedUser) => !newRoomUsersDisplaySet.has(suggestedUser.username)
  );

  const inputIsUsername = suggestedUsers?.some(
    (suggestedUser) => suggestedUser.username === input
  );
  const inputIsNewRoomUser = newRoomUsersDisplaySet.has(input);

  const showFirstOption =
    input && !inputIsUsername && isValidEmail && !inputIsNewRoomUser;
  const showUsernameOptions =
    input && suggestedUsers && suggestedUsers.length > 0;
  const showOptions = showFirstOption || showUsernameOptions;

  const firstOption = showFirstOption ? (
    <AutocompleteOption user={{ email: input }} key={0} isEmail />
  ) : null;

  const usernameOptions = showUsernameOptions
    ? suggestedUsers.map((suggestedUser) => (
        <AutocompleteOption user={suggestedUser} key={suggestedUser.username} />
      ))
    : null;

  const options = showOptions ? (
    <Combobox.Options className="absolute z-10 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
      {firstOption}
      {usernameOptions}
    </Combobox.Options>
  ) : null;

  return (
    <Combobox as="div" value={{ email: input }} onChange={handleOnChange}>
      <div className="relative mt-2">
        <Combobox.Input
          className="w-full rounded-md border-0 bg-white py-1.5 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 placeholder:italic"
          onChange={handleInputChange}
          placeholder="type in username or email address .."
          displayValue={(user: Partial<User>) => user.email || ""}
        />
        {options}
      </div>
    </Combobox>
  );
};
