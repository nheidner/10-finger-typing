import { User } from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/router";
import { Fragment } from "react";
import { Menu, Transition } from "@headlessui/react";
import classNames from "classnames";
import { getAuthenticatedUser, logout } from "@/utils/queries";

const UserMenu = ({ user }: { user?: User }) => {
  const queryClient = useQueryClient();
  const router = useRouter();

  const logoutMutation = useMutation({
    mutationKey: ["logout"],
    mutationFn: logout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["authenticatedUser"] });
      queryClient.removeQueries({ predicate: () => true });

      router.push("/login");
    },
  });

  if (!user) {
    return null;
  }

  const onLogout = async () => {
    logoutMutation.mutate();
  };

  return (
    <Menu as="div" className="relative ml-3">
      <div>
        <Menu.Button className="flex max-w-xs items-center rounded-full text-sm focus:outline-none focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-gray-800">
          <span className="sr-only">Open user menu</span>
          <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-gray-300">
            <span className="text-md font-medium leading-none text-white">
              {user.username[0].toUpperCase()}
            </span>
          </span>
        </Menu.Button>
      </div>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
          <Menu.Item>
            {({ active }) => (
              <Link
                href={`/${user.username}`}
                className={classNames(
                  active ? "bg-gray-100" : "",
                  "block px-4 py-2 text-sm text-gray-700"
                )}
              >
                Profile
              </Link>
            )}
          </Menu.Item>
          <Menu.Item>
            {({ active }) => (
              <button
                onClick={onLogout}
                className={classNames(
                  active ? "bg-gray-100" : "",
                  "block w-full px-4 py-2 text-sm text-gray-700 text-left"
                )}
              >
                Logout
              </button>
            )}
          </Menu.Item>
        </Menu.Items>
      </Transition>
    </Menu>
  );
};

export const Header = () => {
  const { data, isError } = useQuery({
    queryKey: ["authenticatedUser"],
    queryFn: () => getAuthenticatedUser(),
    retry: false,
  });

  const navigation = [
    { name: "Home", href: "/" },
    { name: "Train", href: "/train" },
  ];

  const userIsLoggedIn = !isError && !!data;

  return (
    // Todo: split up into components
    <header className="bg-white mb-11">
      <nav
        className="flex items-center justify-between gap-x-6 py-6"
        aria-label="Global"
      >
        <div className="flex gap-x-12">
          {userIsLoggedIn
            ? navigation.map((item) => (
                <Link
                  key={item.name}
                  href={item.href}
                  className="text-sm font-semibold leading-6 text-gray-900"
                >
                  {item.name}
                </Link>
              ))
            : null}
        </div>
        <div className="flex flex-1 items-center justify-end gap-x-6">
          {userIsLoggedIn ? (
            <UserMenu user={data} />
          ) : (
            <>
              <Link
                href="/login"
                className="hidden lg:block lg:text-sm lg:font-semibold lg:leading-6 lg:text-gray-900"
              >
                Log in
              </Link>
              <Link
                href="/signup"
                className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
              >
                Sign up
              </Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
};
