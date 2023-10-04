import { getScoresByUsername } from "@/utils/queries";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { SortByOption } from "../types";
import { SortBy } from "./SortBy";
import { Score } from "@/types";

export const ScoresList = ({ data }: { data?: Score[] }) => {
  if (!data) {
    return (
      <tbody>
        <tr>
          <td colSpan={4}>
            <div
              className="flex justify-center items-center w-full"
              style={{ height: "522.5px" }}
            >
              Loading ...
            </div>
          </td>
        </tr>
      </tbody>
    );
  }

  return (
    <tbody className="bg-white">
      {data.slice(0, 10).map((score) => {
        const wordsPerMinute = score.wordsPerMinute.toFixed(1);
        const accuracy = `${score.accuracy.toFixed(2)} %`;
        const numberErrors = score.numberErrors;

        return (
          <tr key={score.id} className="even:bg-gray-50">
            <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-3">
              {wordsPerMinute}
            </td>
            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              {accuracy}
            </td>
            <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              {numberErrors}
            </td>
            <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-3">
              <a href="#" className="text-indigo-600 hover:text-indigo-900">
                Edit<span className="sr-only">, {score.id}</span>
              </a>
            </td>
          </tr>
        );
      })}
    </tbody>
  );
};

export const Scores = ({ username }: { username: string }) => {
  const [sortBy, setSortBy] = useState<SortByOption>("recent");

  const sortByParamValues = sortBy === "recent" ? [] : [`${sortBy}.desc`];

  const { data, isLoading } = useQuery({
    queryKey: ["score", username, sortBy],
    queryFn: () => getScoresByUsername(username, { sortBy: sortByParamValues }),
  });

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
      <table className="min-w-full divide-y divide-gray-300">
        <thead>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-3 block"
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

        <ScoresList data={data} />
      </table>
    </section>
  );
};
