import type { User } from "@/types";
import { fetchApi } from "@/utils/fetch";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const getUserByUsername = async (username: string, cookie?: string) => {
  const headers = cookie ? { cookie } : undefined;

  return fetchApi<User>(`/users/${username}`, { headers });
};

const Avatar = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  return (
    <span className="inline-flex h-48 w-48 items-center justify-center rounded-full bg-gray-300">
      <span className="text-8xl font-medium leading-none text-white">
        {user.username[0].toUpperCase()}
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

const Profile = ({ user }: { user?: User }) => {
  if (!user) {
    return null;
  }

  return (
    <section className="flex gap-[10%] border-b border-gray-100 mb-6 pb-8">
      <Avatar user={user} />
      <ProfileData user={user} />
    </section>
  );
};

const Scores = () => {
  return (
    <section className="">
      <div className="px-4 sm:px-0">
        <h3 className="text-base font-semibold leading-7 text-gray-900">
          Your scores
        </h3>
        <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-500">
          Here are your most recent scores
        </p>
      </div>
    </section>
  );
};

const ProfilePage: NextPage<{
  username: string;
  dehydratedState: DehydratedState;
}> = ({ username }) => {
  const { data } = useQuery({
    queryKey: ["user", username],
    queryFn: () => getUserByUsername(username),
  });

  return (
    <>
      <Profile user={data} />
      <Scores />
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

  return {
    username,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
