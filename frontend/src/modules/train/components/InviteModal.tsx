import { ChangeEvent, FC, Fragment, useMemo, useState } from "react";
import { Dialog, Transition, Combobox } from "@headlessui/react";
import {
  PlusCircleIcon,
  AtSymbolIcon,
  XMarkIcon,
} from "@heroicons/react/20/solid";
import classNames from "classnames";
import { useDebouncedUserSearchByUsernamePartial } from "../hooks/use_debounced_user_search_by_username_partial.ts";
import isEmail from "validator/lib/isEmail";
import Link from "next/link.js";
import { Avatar } from "@/components/Avatar";
import { User } from "@/types.js";

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

const UsernameAutoComplete: FC<{
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

const UserList: FC<{
  newRoomUsers: Partial<User>[];
  removeNewRoomUser: (i: number) => void;
}> = ({ newRoomUsers, removeNewRoomUser }) => {
  if (!newRoomUsers.length) {
    return null;
  }

  return (
    <ul className="grid gap-y-6 grid-cols-3 mb-8">
      {newRoomUsers.map((newRoomUser, i) => {
        const userString = newRoomUser.username || newRoomUser.email;
        const hasUsername = !!newRoomUser.username;

        const handleRemoveNewRoomUser = () => {
          removeNewRoomUser(i);
        };

        return (
          <li
            key={userString}
            className="flex flex-col items-center w-full group relative"
          >
            <Avatar
              user={newRoomUser}
              textClassName="text-4xl"
              containerClassName="w-24 h-24"
            />
            <span className="text-center block text-xs mt-1 w-full overflow-hidden px-1">
              {hasUsername ? (
                <Link
                  href={`/${userString}`}
                  target="_blank"
                  className="hover:underline"
                >
                  {userString}
                </Link>
              ) : (
                userString
              )}
            </span>
            <button
              onClick={handleRemoveNewRoomUser}
              type="button"
              className="absolute top-1/2 -translate-y-1/2 opacity-0 group-hover:visible group-hover:opacity-100 transition-opacity duration-150 rounded bg-white px-2 py-1 text-sm font-semibold text-red-700 shadow-sm ring-1 ring-inset ring-red-700 hover:bg-gray-50"
            >
              remove
            </button>
          </li>
        );
      })}
    </ul>
  );
};

const InvitePanel: FC = () => {
  const [newRoomUsers, setNewRoomUsers] = useState<Partial<User>[]>([]);

  const addNewRoomUser = (user: Partial<User>) => {
    setNewRoomUsers((users) => users.concat(user));
  };

  const removeNewRoomUser = (idx: number) => {
    setNewRoomUsers((prevUsers) => prevUsers.filter((_, i) => i !== idx));
  };

  const newRoomUsersDisplaySet = useMemo(
    () =>
      new Set(
        newRoomUsers.map(
          (newRoomUser) => (newRoomUser.username || newRoomUser.email)!
        )
      ),
    [newRoomUsers]
  );

  return (
    <Transition.Child
      as={Fragment}
      enter="ease-out duration-50"
      enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
      enterTo="opacity-100 translate-y-0 sm:scale-100"
      leave="ease-out duration-20"
      leaveFrom="opacity-100 translate-y-0 sm:scale-100"
      leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
    >
      <Dialog.Panel className="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-[31.25rem] sm:p-6">
        <UserList
          newRoomUsers={newRoomUsers}
          removeNewRoomUser={removeNewRoomUser}
        />
        <UsernameAutoComplete
          addNewRoomUser={addNewRoomUser}
          newRoomUsersDisplaySet={newRoomUsersDisplaySet}
        />
      </Dialog.Panel>
    </Transition.Child>
  );
};

export const InviteModal: FC<{
  isOpen: boolean;
  setOpen: (open: boolean) => void;
}> = ({ isOpen, setOpen }) => {
  return (
    <Transition.Root show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-10" onClose={setOpen}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-50"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-out duration-20"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black bg-opacity-25 transition-opacity backdrop-blur-sm" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <InvitePanel />
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
};
