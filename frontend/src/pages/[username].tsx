import { Profile } from "@/modules/profile/components/Profile";
import { Scores } from "@/modules/profile/components/Scores";
import { getScoresByUsername, getUserByUsername } from "@/utils/queries";
import { DehydratedState, QueryClient, dehydrate } from "@tanstack/react-query";
import { NextPage } from "next";

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
  await queryClient.prefetchQuery(["score", username, "recent"], () =>
    getScoresByUsername(username, { cookie })
  );

  return {
    username,
    dehydratedState: dehydrate(queryClient),
  };
};

export default ProfilePage;
