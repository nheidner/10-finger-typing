import { getScoresByUsername } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { SortByOption } from "../types";
import { SortBy } from "./SortBy";
import { Score } from "@/types";
import { LoadingOverlay } from "@/components/LoadingOverlay";
import { ScoresTable } from "@/components/ScoresTable";

export const Scores = ({ username }: { username: string }) => {
  const [sortBy, setSortBy] = useState<SortByOption>("recent");
  const sortByParamValues = sortBy === "recent" ? [] : [`${sortBy}.desc`];

  const dataRef = useRef<Score[] | undefined>(undefined);

  const { data, isLoading } = useQuery({
    queryKey: ["score", username, sortBy],
    queryFn: () => getScoresByUsername(username, { sortBy: sortByParamValues }),
  });

  if (data) {
    dataRef.current = data;
  }

  return (
    <section>
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
      <LoadingOverlay isLoading={isLoading}>
        <ScoresTable scores={dataRef.current} />
      </LoadingOverlay>
    </section>
  );
};
