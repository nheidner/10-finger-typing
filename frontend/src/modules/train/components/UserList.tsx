import { FC } from "react";
import Link from "next/link";
import { Avatar } from "@/components/Avatar";
import { User } from "@/types";
import { Transition } from "@headlessui/react";

export const UserList: FC<{
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

        const userDisplay = hasUsername ? (
          <Link
            href={`/${userString}`}
            target="_blank"
            className="hover:underline"
          >
            {userString}
          </Link>
        ) : (
          userString
        );

        return (
          <Transition
            key={userString}
            className="flex flex-col items-center w-full group relative"
            as="li"
            appear={true}
            show={true}
            enter="transition-opacity duration-150"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity duration-150"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <Avatar
              user={newRoomUser}
              textClassName="text-4xl"
              containerClassName="w-24 h-24"
            />
            <span className="text-center block text-xs mt-1 w-full overflow-hidden px-1">
              {userDisplay}
            </span>
            <button
              onClick={handleRemoveNewRoomUser}
              type="button"
              className="absolute top-1/2 -translate-y-1/2 opacity-0 group-hover:visible group-hover:opacity-100 transition-opacity duration-150 rounded bg-white px-2 py-1 text-sm font-semibold text-red-700 shadow-sm ring-1 ring-inset ring-red-700 hover:bg-gray-50"
            >
              remove
            </button>
          </Transition>
        );
      })}
    </ul>
  );
};
