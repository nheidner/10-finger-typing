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
      <div>{JSON.stringify(data)}</div>
    </>
  );
};

ProfilePage.getInitialProps = async (ctx) => {
  const { username } = ctx.query as { username: string };
  const queryClient = new QueryClient();
  const { cookie } = ctx.req?.headers || {};

  await queryClient.prefetchQuery(["user", username], () =>
    getUserByUsername(username, cookie)
  );

  return {
    username,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
