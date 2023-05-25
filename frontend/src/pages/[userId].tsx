import type { User } from "@/types";
import { fetchApi } from "@/utils/fetch";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const getUserById = async (id: string, cookie?: string) => {
  const headers = cookie ? { cookie } : undefined;

  return fetchApi<User>(`/users/${id}`, { headers });
};

const ProfilePage: NextPage<{
  userId: string;
  dehydratedState: DehydratedState;
}> = ({ userId }) => {
  const { data } = useQuery({
    queryKey: ["user", userId],
    queryFn: () => getUserById(userId),
  });

  return (
    <>
      <div>{JSON.stringify(data)}</div>
    </>
  );
};

ProfilePage.getInitialProps = async (ctx) => {
  const { userId } = ctx.query as { userId: string };
  const queryClient = new QueryClient();
  const { cookie } = ctx.req?.headers || {};

  await queryClient.prefetchQuery(["user", userId], () =>
    getUserById(userId, cookie)
  );

  return {
    userId,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
