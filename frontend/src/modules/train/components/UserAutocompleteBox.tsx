import { ChangeEvent, FC, useState } from "react";
import { Combobox } from "@headlessui/react";
import { PlusCircleIcon } from "@heroicons/react/20/solid";
import classNames from "classnames";
import { useDebouncedUserSearchByUsernamePartial } from "../hooks/use_debounced_user_search_by_username_partial.ts";
import isEmail from "validator/lib/isEmail";
import { User } from "@/types";

const Option: FC<{ user: Partial<User> }> = ({ user }) => {
  const userDisplay = user.username || user.email;

  return (
    <Combobox.Option
      key={userDisplay}
      value={user}
      className={({ active }) =>
        classNames(
          "relative cursor-pointer select-none py-2 px-3",
          active ? "bg-indigo-600 text-white" : "text-gray-900"
        )
      }
    >
      {({ active }) => (
        <div className="flex justify-between items-center">
          <div className="flex justify-start items-center">
            <span className="truncate">{userDisplay}</span>
          </div>
          <PlusCircleIcon
            className={classNames(
              "h-4 w-4",
              active ? "text-white" : "text-indigo-600"
            )}
          />
        </div>
      )}
    </Combobox.Option>
  );
};

export const UserAutocompleteBox: FC<{
  addNewRoomUser: (user: Partial<User>) => void;
  newRoomUsersDisplaySet: Set<string>;
}> = ({ addNewRoomUser, newRoomUsersDisplaySet }) => {
  const [input, setInput] = useState("");

  const { users, debouncedFetchUsers } =
    useDebouncedUserSearchByUsernamePartial(200);

  const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;

    debouncedFetchUsers(value);
    setInput(value);
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
    <Option user={{ email: input }} key={0} />
  ) : null;

  const usernameOptions = showUsernameOptions
    ? suggestedUsers.map((suggestedUser) => (
        <Option user={suggestedUser} key={suggestedUser.username} />
      ))
    : null;

  const options = showOptions ? (
    <Combobox.Options className="absolute z-10 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
      {firstOption}
      {usernameOptions}
    </Combobox.Options>
  ) : null;

  return (
    <Combobox as="div" value={null} onChange={addNewRoomUser}>
      <div className="relative mt-2">
        <Combobox.Input
          className="w-full rounded-md border-0 bg-white py-1.5 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 placeholder:italic"
          onChange={handleInputChange}
          placeholder="type in username or email address .."
        />
        {options}
      </div>
    </Combobox>
  );
};
