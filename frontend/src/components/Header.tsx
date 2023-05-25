import { User } from "@/types";
import { fetchApi } from "@/utils/fetch";
import { Dialog } from "@headlessui/react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/router";
// import { Bars3Icon, XMarkIcon } from "@radix-ui/react-icons";
import { useState } from "react";

const getLoggedInUser = async () => fetchApi<User>("/user");

const logout = async () =>
  fetchApi<string>("/users/logout", { method: "POST" });

export const Header = () => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const queryClient = useQueryClient();

  const router = useRouter();

  const { data, isError } = useQuery({
    queryKey: ["loggedInUser"],
    queryFn: () => getLoggedInUser(),
    retry: false,
  });

  const logoutMutation = useMutation({
    mutationKey: ["logout"],
    mutationFn: logout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["loggedInUser"] });
      queryClient.removeQueries({ predicate: () => true });
    },
  });

  const navigation = [
    { name: "Home", href: "/" },
    { name: "Profile", href: `/${data?.id}` },
  ];

  const onLogout = async () => {
    logoutMutation.mutate();
  };

  const userIsLoggedIn = !isError && data;

  return (
    // Todo: split up into components
    <header className="bg-white">
      <nav
        className="mx-auto flex max-w-7xl items-center justify-between gap-x-6 p-6 lg:px-8"
        aria-label="Global"
      >
        <div className="hidden lg:flex lg:gap-x-12">
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
            <button
              onClick={onLogout}
              className="hidden lg:block lg:text-sm lg:font-semibold lg:leading-6 lg:text-gray-900"
            >
              Logout
            </button>
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
        {/* <div className="flex lg:hidden">
          <button
            type="button"
            className="-m-2.5 inline-flex items-center justify-center rounded-md p-2.5 text-gray-700"
            onClick={() => setMobileMenuOpen(true)}
          >
            <span className="sr-only">Open main menu</span>
            <Bars3Icon className="h-6 w-6" aria-hidden="true" />
          </button>
        </div> */}
      </nav>
    </header>
  );
};
