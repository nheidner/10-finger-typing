import type { Score, User } from "@/types";
import { fetchApi } from "@/utils/fetch";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import classNames from "classnames";
import { NextPage } from "next";
import { useState } from "react";

const sortByOptions = ["recent", "accuracy", "errors"] as const;

type SortByOption = (typeof sortByOptions)[number];

const getUserByUsername = async (username: string, cookie?: string) => {
  const queryParams = `username=${encodeURIComponent(username)}`;
  const queryString = `?${queryParams}`;

  const headers = cookie ? { cookie } : undefined;

  const users = await fetchApi<User[]>(`/users${queryString}`, { headers });
  return users[0];
};

const getScoresByUsername = async (
  username: string,
  { cookie, sortBy }: { cookie?: string; sortBy?: string[] }
) => {
  const queryParams = sortBy
    ?.map((sortByValue) => `sort_by=${encodeURIComponent(sortByValue)}`)
    .concat(`username=${encodeURIComponent(username)}`)
    .join("&");
  const queryString = queryParams ? `?${queryParams}` : "";

  const headers = cookie ? { cookie } : undefined;

  return fetchApi<Score[]>(`/scores${queryString}`, {
    headers,
  });
};

const Avatar = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  return (
    <span className="inline-flex h-48 w-48 items-center justify-center rounded-full bg-gray-300">
      <span className="text-8xl font-medium leading-none text-white">
        {user.username?.[0].toUpperCase()}
      </span>
    </span>
  );
};

const ProfileData = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  const data = [
    { name: "Username", value: user.username },
    { name: "Email", value: user.email },
    { name: "Full name", value: `${user.firstName} ${user.lastName}` },
  ];

  return (
    <div className="flex-1">
      <div className="px-4 sm:px-0">
        <h3 className="text-base font-semibold leading-7 text-gray-900">
          Personal Information
        </h3>
        {/* <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-500">
          Personal details and application.
        </p> */}
      </div>
      <div className="mt-6 border-t border-gray-100">
        {/* <dl className="divide-y divide-gray-100"> */}
        <dl className="">
          {data.map((item) => {
            return (
              <div
                key={item.name}
                className="px-4 py-6 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-0"
              >
                <dt className="text-sm font-medium leading-6 text-gray-900">
                  {item.name}
                </dt>
                <dd className="mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0">
                  {item.value}
                </dd>
              </div>
            );
          })}
        </dl>
      </div>
    </div>
  );
};

const Profile = ({ username }: { username: string }) => {
  const { data } = useQuery({
    queryKey: ["user", username],
    queryFn: () => getUserByUsername(username),
  });

  if (!data) {
    return null;
  }

  return (
    <section className="flex gap-[10%] border-b border-gray-100 mb-6 pb-8">
      <Avatar user={data} />
      <ProfileData user={data} />
    </section>
  );
};

const SortBy = ({
  sortBy,
  setSortBy,
}: {
  sortBy: SortByOption;
  setSortBy: (sortByOption: SortByOption) => void;
}) => {
  const selectOption = (sortByOption: SortByOption) => () =>
    setSortBy(sortByOption);

  return (
    <div className="flex items-center">
      <h4 className="text-sm font-semibold leading-7 text-gray-700">
        Sort by:
      </h4>
      {sortByOptions
        .map((sortByOption, index) => {
          const isLast = index === sortByOptions.length - 1;
          const isActive = sortByOption === sortBy;

          const option = (
            <span
              onClick={selectOption(sortByOption)}
              key={sortByOption}
              className={classNames(
                "inline-block mx-2 text-sm cursor-pointer",
                isActive && "font-semibold text-indigo-400"
              )}
            >
              {sortByOption}
            </span>
          );

          if (isLast) {
            return option;
          }

          return [
            option,
            <span key={index} className="text-gray-500">
              |
            </span>,
          ];
        })
        .flat()}
    </div>
  );
};

const Scores = ({ username }: { username: string }) => {
  const [sortBy, setSortBy] = useState<SortByOption>("recent");

  const sortByParamValues = sortBy === "recent" ? [] : [`${sortBy}.desc`];

  const { data } = useQuery({
    queryKey: ["score", username, sortBy],
    queryFn: () => getScoresByUsername(username, { sortBy: sortByParamValues }),
  });

  if (!data) {
    return null;
  }

  return (
    <section className="">
      <div className="px-4 sm:px-0 mb-10 flex justify-between items-center">
        <div>
          <h3 className="text-base font-semibold leading-7 text-gray-900">
            Your scores
          </h3>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-500">
            Here are your most recent scores
          </p>
        </div>
        <SortBy sortBy={sortBy} setSortBy={setSortBy} />
      </div>
      <table className="min-w-full divide-y divide-gray-300">
        <thead>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-3"
            >
              Words Per Minute
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
            >
              Accuracy
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
            >
              Number of Errors
            </th>
            <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-3">
              <span className="sr-only">Edit</span>
            </th>
          </tr>
        </thead>
        <tbody className="bg-white">
          {data.map((score) => (
            <tr key={score.id} className="even:bg-gray-50">
              <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-3">
                {score.wordsPerMinute}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                {score.accuracy}
              </td>
              <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                {score.numberErrors}
              </td>
              <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-3">
                <a href="#" className="text-indigo-600 hover:text-indigo-900">
                  Edit<span className="sr-only">, {score.id}</span>
                </a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
};

const ProfilePage: NextPage<{
  username: string;
  dehydratedState: DehydratedState;
}> = ({ username }) => {
  return (
    <>
      <Profile username={username} />
      <Scores username={username} />
    </>
  );
};

ProfilePage.getInitialProps = async (ctx) => {
  const { username } = ctx.query as { username: string };
  const { cookie } = ctx.req?.headers || {};
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery(["user", username], () =>
    getUserByUsername(username, cookie)
  );
  await queryClient.prefetchQuery(["score", username], () =>
    getScoresByUsername(username, { cookie })
  );

  return {
    username,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
