import { FC } from "react";
import { Combobox } from "@headlessui/react";
import { EnvelopeIcon, PlusIcon } from "@heroicons/react/20/solid";
import classNames from "classnames";
import { User } from "@/types";
import { Avatar } from "@/components/Avatar";

export const AutocompleteOption: FC<{
  user: Partial<User>;
  isEmail?: boolean;
}> = ({ user, isEmail }) => {
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
      {({ active }) => {
        const leadingIcon = isEmail ? (
          <EnvelopeIcon
            className={classNames(
              "h-6 w-6 mr-2",
              active ? "text-white" : "text-indigo-600"
            )}
          />
        ) : (
          <Avatar
            user={user}
            textClassName="text-xs"
            containerClassName="h-6 w-6 mr-2"
          />
        );

        return (
          <div className="flex justify-between items-center">
            {leadingIcon}
            <span className="truncate flex-1">{userDisplay}</span>
            <PlusIcon
              className={classNames(
                "h-4 w-4",
                active ? "text-white" : "text-indigo-600"
              )}
            />
          </div>
        );
      }}
    </Combobox.Option>
  );
};
