import { ChangeEvent, FC, Fragment, useState } from "react";
import { Dialog, Transition, Combobox } from "@headlessui/react";
import { CheckIcon } from "@heroicons/react/20/solid";
import classNames from "classnames";
import { useDebouncedUserSearchByUsernamePartial } from "../hooks/use_debounced_user_search_by_username_partial.ts";

const Option: FC<{ content: string }> = ({ content }) => {
  return (
    <Combobox.Option
      key={content}
      value={content}
      className={({ active }) =>
        classNames(
          "relative cursor-default select-none py-2 pl-3 pr-9",
          active ? "bg-indigo-600 text-white" : "text-gray-900"
        )
      }
    >
      {({ active, selected }) => (
        <>
          <div className="flex">
            <span
              className={classNames("truncate", selected && "font-semibold")}
            >
              {content}
            </span>
          </div>

          {selected && (
            <span
              className={classNames(
                "absolute inset-y-0 right-0 flex items-center pr-4",
                active ? "text-white" : "text-indigo-600"
              )}
            >
              <CheckIcon className="h-5 w-5" aria-hidden="true" />
            </span>
          )}
        </>
      )}
    </Combobox.Option>
  );
};

const UsernameAutoComplete: FC<{
  handleSelectNewRoomUser: (user: string) => void;
}> = ({ handleSelectNewRoomUser }) => {
  const [input, setInput] = useState("");

  const { users, debouncedFetchUsers } =
    useDebouncedUserSearchByUsernamePartial();

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;

    debouncedFetchUsers(value);
    setInput(value);
  };

  const usernames = users?.map((user) => user.username);

  const showFirstOption = input && !usernames?.includes(input);
  const showUsernameOptions = input && usernames && usernames.length > 0;
  const showOptions = showFirstOption || showUsernameOptions;

  const firstOption = showFirstOption ? (
    <Option content={input} key={0} />
  ) : null;

  const usernameOptions = showUsernameOptions
    ? usernames.map((username) => <Option content={username} key={username} />)
    : null;

  const options = showOptions ? (
    <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
      {firstOption}
      {usernameOptions}
    </Combobox.Options>
  ) : null;

  return (
    <Combobox as="div" value={null} onChange={handleSelectNewRoomUser}>
      <div className="relative mt-2">
        <Combobox.Input
          className="w-full rounded-md border-0 bg-white py-1.5 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 placeholder:italic"
          onChange={handleChange}
          placeholder="type in username or email address .."
        />
        {options}
      </div>
    </Combobox>
  );
};

export const InviteModal: FC<{
  isOpen: boolean;
  setOpen: (open: boolean) => void;
}> = ({ isOpen, setOpen }) => {
  const [newRoomUsers, setNewUsers] = useState<string[]>([]);

  const handleSelectNewRoomUser = (user: string) => {
    setNewUsers((users) => users.concat(user));
  };

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
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 w-screen overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-50"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-out duration-20"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <Dialog.Panel className="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6">
                <UsernameAutoComplete
                  handleSelectNewRoomUser={handleSelectNewRoomUser}
                />
                <div>
                  {newRoomUsers.map((user) => {
                    return <p key={user}>{user}</p>;
                  })}
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
};
