import { Score } from "@/types";
import { FC } from "react";

const ScoresList = ({ scores }: { scores: Score[] }) => {
  return (
    <tbody className="bg-white">
      {scores.map((score) => {
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

export const ScoresTable: FC<{ scores?: Score[] }> = ({ scores }) => {
  if (!scores) {
    return null;
  }

  return (
    <table className="min-w-full divide-y divide-gray-300 table-fixed">
      <thead className="z-20 bg-white relative">
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
      <ScoresList scores={scores} />
    </table>
  );
};
