import type { Book, User } from "@/types";
import { fetchApi } from "@/utils/fetch";
import { getApiUrl } from "@/utils/get_api_url";
import {
  DehydratedState,
  QueryClient,
  dehydrate,
  useQuery,
} from "@tanstack/react-query";
import { NextPage } from "next";

const getUserById = async (id: string) => fetchApi<User>(`/users/${id}`);

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

  await queryClient.prefetchQuery(["user", userId], () => getUserById(userId));

  return {
    userId,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
